package hub

import (
	"sync"
	"time"

	"github.com/google/uuid"

	"gowiki/internal/collab/crdt"
	"gowiki/internal/collab/lock"
	"gowiki/internal/collab/protocol"
	"gowiki/internal/collab/snapshot"
	"gowiki/internal/logger"
	"gowiki/internal/pkg/timeutil"
)

type persistFunc func(docID uuid.UUID, text, state string, ops []crdt.Op)

type snapshotFunc func(docID uuid.UUID, reason string)

type client struct {
	id     string
	userID string
	name   string
	color  string
	cursor int
	site   uint64
	send   chan protocol.Envelope
}

type Room struct {
	id         uuid.UUID
	mu         sync.Mutex
	doc        *crdt.Doc
	clients    map[string]*client
	locks      *lock.Store
	persist    persistFunc
	snapshot   snapshotFunc
	dirtyRunes int
	lastChange time.Time
	lastSnap   time.Time
	pendingOps []crdt.Op
}

func newRoom(id uuid.UUID, state string, locks *lock.Store, persist persistFunc, snapshot snapshotFunc) *Room {
	d := crdt.NewDoc(1)
	if snap, err := crdt.ParseSnapshot(state); err == nil && len(snap.Atoms) > 0 {
		d.LoadSnapshot(snap)
	}
	now := timeutil.Now()
	return &Room{
		id: id, doc: d, clients: map[string]*client{},
		locks: locks, persist: persist, snapshot: snapshot,
		lastChange: now, lastSnap: now,
	}
}

func (r *Room) add(c *client) {
	r.mu.Lock()
	r.clients[c.id] = c
	text := r.doc.Text()
	clock := r.doc.Clock()
	locks := r.lockStates()
	users := r.presences()
	r.mu.Unlock()
	snap := r.doc.Snapshot()
	c.send <- protocol.Envelope{
		Type: protocol.TypeSnapshot, Text: text, Clock: clock, SiteID: c.site,
		Users: users, Locks: locks, Atoms: snap.Atoms,
	}
	r.broadcast(protocol.Envelope{Type: protocol.TypePresence, Users: users}, "")
}

func (r *Room) remove(id string) {
	r.mu.Lock()
	c, ok := r.clients[id]
	if ok {
		delete(r.clients, id)
		dropped := r.locks.ReleaseAll(r.id.String(), c.userID)
		users := r.presences()
		r.mu.Unlock()
		for _, rec := range dropped {
			r.broadcast(protocol.Envelope{
				Type: protocol.TypeLock, Paragraph: rec.ParagraphID, Action: "release",
			}, "")
		}
		r.broadcast(protocol.Envelope{Type: protocol.TypePresence, Users: users}, "")
		return
	}
	r.mu.Unlock()
}

func (r *Room) apply(from *client, op crdt.Op) error {
	r.mu.Lock()
	if err := r.doc.Apply(op); err != nil {
		r.mu.Unlock()
		return err
	}
	r.pendingOps = append(r.pendingOps, op)
	if op.Type == crdt.OpInsert {
		r.dirtyRunes++
	}
	r.lastChange = timeutil.Now()
	r.mu.Unlock()
	r.broadcast(protocol.Envelope{Type: protocol.TypeOp, Op: &op, SiteID: from.site}, from.id)
	return nil
}

func (r *Room) setPresence(from *client, cursor int, color string) {
	r.mu.Lock()
	from.cursor = cursor
	if color != "" {
		from.color = color
	}
	users := r.presences()
	r.mu.Unlock()
	r.broadcast(protocol.Envelope{Type: protocol.TypePresence, Users: users}, "")
}

func (r *Room) handleLock(from *client, paragraph, action string) protocol.Envelope {
	docKey := r.id.String()
	switch action {
	case "acquire":
		rec, ok := r.locks.Acquire(docKey, paragraph, from.userID, from.name)
		env := protocol.Envelope{
			Type: protocol.TypeLock, Paragraph: paragraph, Action: action,
			Holder: rec.HolderID, Until: timeutil.Format(rec.Until),
		}
		if !ok {
			env.Code = "PARAGRAPH_LOCKED"
			env.Message = rec.HolderName + " 正在编辑该段落"
			return env
		}
		r.broadcast(env, "")
		return env
	case "heartbeat":
		rec, ok := r.locks.Heartbeat(docKey, paragraph, from.userID)
		if !ok {
			return protocol.Envelope{Type: protocol.TypeError, Code: "LOCK_LOST", Message: "锁已失效"}
		}
		return protocol.Envelope{Type: protocol.TypeLock, Paragraph: paragraph, Action: action, Holder: rec.HolderID, Until: timeutil.Format(rec.Until)}
	default:
		r.locks.Release(docKey, paragraph, from.userID)
		env := protocol.Envelope{Type: protocol.TypeLock, Paragraph: paragraph, Action: "release"}
		r.broadcast(env, from.id)
		return env
	}
}

func (r *Room) flushIfNeeded(force bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := timeutil.Now()
	dec := snapshot.Evaluate(now.Sub(r.lastChange), r.dirtyRunes, now.Sub(r.lastSnap), len(r.pendingOps), force)
	if !dec.PersistOps && !dec.AutoL2 {
		return
	}
	text := r.doc.Text()
	state, _ := r.doc.Snapshot().Marshal()
	ops := append([]crdt.Op{}, r.pendingOps...)
	r.pendingOps = nil
	r.dirtyRunes = 0
	r.lastSnap = now
	if r.persist != nil && dec.PersistOps {
		go r.persist(r.id, text, state, ops)
	}
	if r.snapshot != nil && dec.AutoL2 {
		go r.snapshot(r.id, dec.Reason)
	}
}

func (r *Room) empty() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.clients) == 0
}

func (r *Room) broadcast(env protocol.Envelope, except string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, c := range r.clients {
		if id == except {
			continue
		}
		select {
		case c.send <- env:
		default:
			logger.L().Warn("client send buffer full", "client", id)
		}
	}
}

func (r *Room) presences() []protocol.Presence {
	out := make([]protocol.Presence, 0, len(r.clients))
	for _, c := range r.clients {
		out = append(out, protocol.Presence{
			UserID: c.userID, Name: c.name, Color: c.color, Cursor: c.cursor,
		})
	}
	return out
}

func (r *Room) lockStates() []protocol.LockState {
	recs := r.locks.List(r.id.String())
	out := make([]protocol.LockState, 0, len(recs))
	for _, rec := range recs {
		out = append(out, protocol.LockState{
			ParagraphID: rec.ParagraphID, Holder: rec.HolderID,
			HolderName: rec.HolderName, Until: timeutil.Format(rec.Until),
		})
	}
	return out
}

