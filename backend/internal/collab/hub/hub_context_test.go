package hub

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gowiki/internal/collab/lock"
	"gowiki/internal/collab/protocol"
)

type acceptingParser struct{}

func (acceptingParser) Parse(string) (string, string, string, error) {
	return "user-1", "Alice", "#123456", nil
}

func TestWebSocketRemainsUsableWhileIdle(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	docID := uuid.New()
	locks := lock.New(time.Minute)
	room := newRoom(docID, "", locks, nil, nil)
	h := &Hub{
		rooms:  map[string]*Room{docID.String(): room},
		locks:  locks,
		parser: acceptingParser{},
	}
	h.sites.Store(100)

	router := gin.New()
	router.GET("/ws", h.HandleWS)
	server := httptest.NewServer(router)
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?token=valid&documentId=" + docID.String()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.CloseNow()

	var snapshot protocol.Envelope
	if err := readEnvelope(ctx, conn, &snapshot); err != nil {
		t.Fatalf("read initial snapshot: %v", err)
	}
	if snapshot.Type != protocol.TypeSnapshot {
		t.Fatalf("initial message type = %q, want %q", snapshot.Type, protocol.TypeSnapshot)
	}

	time.Sleep(250 * time.Millisecond)
	ping, err := json.Marshal(protocol.Envelope{Type: protocol.TypePing})
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Write(ctx, websocket.MessageText, ping); err != nil {
		t.Fatalf("write ping after idle period: %v", err)
	}

	for {
		var env protocol.Envelope
		if err := readEnvelope(ctx, conn, &env); err != nil {
			t.Fatalf("read pong after idle period: %v", err)
		}
		if env.Type == protocol.TypePong {
			break
		}
	}
}

func readEnvelope(ctx context.Context, conn *websocket.Conn, env *protocol.Envelope) error {
	_, data, err := conn.Read(ctx)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, env)
}
