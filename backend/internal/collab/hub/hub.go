package hub

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gowiki/internal/collab/crdt"
	"gowiki/internal/collab/lock"
	"gowiki/internal/collab/protocol"
	"gowiki/internal/logger"
	"gowiki/internal/model"
	"gowiki/internal/pkg/timeutil"
	"gowiki/internal/repository"
	"gowiki/internal/service"
)

type TokenParser interface {
	Parse(token string) (userID, name, color string, err error)
}

type Hub struct {
	mu       sync.Mutex
	rooms    map[string]*Room
	locks    *lock.Store
	docs     *repository.DocumentRepo
	ops      *repository.OpRepo
	versions *service.VersionService
	parser   TokenParser
	sites    atomic.Uint64
}

func New(docs *repository.DocumentRepo, ops *repository.OpRepo, versions *service.VersionService, parser TokenParser, lockTimeout time.Duration) *Hub {
	h := &Hub{
		rooms:    map[string]*Room{},
		locks:    lock.New(lockTimeout),
		docs:     docs,
		ops:      ops,
		versions: versions,
		parser:   parser,
	}
	h.sites.Store(100)
	go h.loop()
	return h
}

func (h *Hub) loop() {
	t := time.NewTicker(10 * time.Second)
	for range t.C {
		h.mu.Lock()
		for _, r := range h.rooms {
			r.flushIfNeeded(false)
		}
		h.mu.Unlock()
		_, _ = h.ops.PurgeBefore(timeutil.Now().Add(-24 * time.Hour))
	}
}

func (h *Hub) room(id uuid.UUID) (*Room, error) {
	key := id.String()
	h.mu.Lock()
	if r, ok := h.rooms[key]; ok {
		h.mu.Unlock()
		return r, nil
	}
	h.mu.Unlock()
	doc, err := h.docs.ByID(id)
	if err != nil {
		return nil, err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if r, ok := h.rooms[key]; ok {
		return r, nil
	}
	r := newRoom(id, doc.CRDTState, h.locks, h.persist, h.autoSnap)
	if doc.CRDTState == "" && doc.ContentMD != "" {
		_, _ = r.doc.LocalInsert(0, doc.ContentMD)
	}
	h.rooms[key] = r
	return r, nil
}

func (h *Hub) persist(docID uuid.UUID, text, state string, ops []crdt.Op) {
	d, err := h.docs.ByID(docID)
	if err != nil {
		return
	}
	d.ContentMD = text
	d.CRDTState = state
	d.UpdatedAt = timeutil.Now()
	_ = h.docs.Update(d)
	for _, op := range ops {
		raw, _ := json.Marshal(op)
		_ = h.ops.Append(&model.DocumentOp{
			DocumentID: docID, SiteID: op.ID.Site, Clock: op.ID.Clock,
			OpJSON: string(raw), CreatedAt: timeutil.Now(),
		})
	}
}

func (h *Hub) autoSnap(docID uuid.UUID, reason string) {
	system, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
	_, _ = h.versions.AutoSnapshot(system, docID, "自动快照·"+reason)
}

func (h *Hub) HandleWS(c *gin.Context) {
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		logger.L().Warn("ws accept failed", "err", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "bye")

	ctx := c.Request.Context()
	token := c.Query("token")
	userID, name, color, err := h.parser.Parse(token)
	if err != nil {
		_ = writeJSON(ctx, conn, protocol.Envelope{Type: protocol.TypeError, Code: "UNAUTHORIZED", Message: "未登录"})
		return
	}
	docID, err := uuid.Parse(c.Query("documentId"))
	if err != nil {
		_ = writeJSON(ctx, conn, protocol.Envelope{Type: protocol.TypeError, Code: "BAD_REQUEST", Message: "缺少 documentId"})
		return
	}
	room, err := h.room(docID)
	if err != nil {
		_ = writeJSON(ctx, conn, protocol.Envelope{Type: protocol.TypeError, Code: "NOT_FOUND", Message: "文档不存在"})
		return
	}

	// Atomically reserve a globally-unique site id. The previous code did
	// h.sites.Load()+1 followed by h.sites.Store() as two separate atomic
	// ops; between them it called room.add(cl) (which takes the room mutex,
	// builds a snapshot and broadcasts presence), opening a wide window in
	// which two concurrent joins read the same counter, compute the same
	// nextSite, and join with identical site ids. Both clients then start
	// from the same initial snapshot/clock, so their first ops collide on
	// (site, clock) and the CRDT's idempotency map silently drops one side
	// — exactly the reported "beta" content loss. Add(1) is a single atomic
	// read-modify-write, so every join observes a distinct value.
	nextSite := h.sites.Add(1)
	cl := &client{
		id: uuid.New().String(), userID: userID, name: name, color: color,
		site: nextSite, send: make(chan protocol.Envelope, 32),
	}
	room.add(cl)
	defer func() {
		room.remove(cl.id)
		room.flushIfNeeded(true)
	}()

	go func() {
		for env := range cl.send {
			if err := writeJSON(ctx, conn, env); err != nil {
				return
			}
		}
	}()

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return
		}
		var env protocol.Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			_ = writeJSON(ctx, conn, protocol.Envelope{Type: protocol.TypeError, Code: "VALIDATION", Message: "消息格式错误"})
			continue
		}
		switch env.Type {
		case protocol.TypePing:
			_ = writeJSON(ctx, conn, protocol.Envelope{Type: protocol.TypePong})
		case protocol.TypeOp:
			if env.Op == nil {
				_ = writeJSON(ctx, conn, protocol.Envelope{Type: protocol.TypeError, Code: "VALIDATION", Message: "缺少 op"})
				continue
			}
			if err := protocol.ValidateEnvelope(env); err != nil && env.Type == protocol.TypeOp {
			_ = writeJSON(ctx, conn, protocol.NewError("VALIDATION", err.Error()))
			continue
		}
		if err := room.apply(cl, *env.Op); err != nil {
				_ = writeJSON(ctx, conn, protocol.Envelope{Type: protocol.TypeError, Code: "VALIDATION", Message: err.Error()})
			}
		case protocol.TypePresence:
			room.setPresence(cl, env.Cursor, env.Color)
		case protocol.TypeLock:
			resp := room.handleLock(cl, env.Paragraph, env.Action)
			_ = writeJSON(ctx, conn, resp)
		case protocol.TypeJoin:
			// already joined
		default:
			if env.Type != "" {
				_ = writeJSON(ctx, conn, protocol.Envelope{Type: protocol.TypeError, Code: "VALIDATION", Message: "未知消息类型"})
			}
		}
	}
}

func writeJSON(ctx context.Context, conn *websocket.Conn, env protocol.Envelope) error {
	b, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, b)
}

func (h *Hub) AuthParser(fn TokenParser) { h.parser = fn }
