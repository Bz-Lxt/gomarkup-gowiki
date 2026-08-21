package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gowiki/internal/collab/lock"
	"gowiki/internal/collab/protocol"
)

type memoryAddr string

func (a memoryAddr) Network() string { return string(a) }
func (a memoryAddr) String() string  { return string(a) }

type memoryListener struct {
	connections chan net.Conn
	closed      chan struct{}
	closeOnce   sync.Once
}

func newMemoryListener() *memoryListener {
	return &memoryListener{
		connections: make(chan net.Conn),
		closed:      make(chan struct{}),
	}
}

func (l *memoryListener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.connections:
		return conn, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}

func (l *memoryListener) Close() error {
	l.closeOnce.Do(func() { close(l.closed) })
	return nil
}

func (l *memoryListener) Addr() net.Addr { return memoryAddr("memory") }

func (l *memoryListener) DialContext(ctx context.Context, _, _ string) (net.Conn, error) {
	client, server := net.Pipe()
	select {
	case l.connections <- server:
		return client, nil
	case <-ctx.Done():
		_ = client.Close()
		_ = server.Close()
		return nil, ctx.Err()
	case <-l.closed:
		_ = client.Close()
		_ = server.Close()
		return nil, net.ErrClosed
	}
}

type concurrentJoinParser struct {
	total   int
	mu      sync.Mutex
	waiting int
	release chan struct{}
}

func newConcurrentJoinParser(total int) *concurrentJoinParser {
	return &concurrentJoinParser{total: total, release: make(chan struct{})}
}

func (p *concurrentJoinParser) Parse(token string) (string, string, string, error) {
	p.mu.Lock()
	p.waiting++
	if p.waiting == p.total {
		close(p.release)
	}
	p.mu.Unlock()
	<-p.release
	return token, token, "#3366ff", nil
}

func TestConcurrentWebSocketJoinsReceiveUniqueSiteIDs(t *testing.T) {
	const clientCount = 16

	docID := uuid.New()
	locks := lock.New(time.Minute)
	h := &Hub{
		rooms:  make(map[string]*Room),
		locks:  locks,
		parser: newConcurrentJoinParser(clientCount),
	}
	h.rooms[docID.String()] = newRoom(docID, "", locks, nil, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", h.HandleWS)
	listener := newMemoryListener()
	server := &http.Server{Handler: router}
	go func() { _ = server.Serve(listener) }()
	defer server.Close()
	transport := &http.Transport{DialContext: listener.DialContext}
	defer transport.CloseIdleConnections()
	httpClient := &http.Client{Transport: transport}

	type result struct {
		site uint64
		err  error
	}
	results := make(chan result, clientCount)
	releaseClients := make(chan struct{})
	var clients sync.WaitGroup

	for i := 0; i < clientCount; i++ {
		clients.Add(1)
		go func(i int) {
			defer clients.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			url := fmt.Sprintf("ws://gowiki.test/ws?token=user-%d&documentId=%s", i, docID)
			conn, _, err := websocket.Dial(ctx, url, &websocket.DialOptions{HTTPClient: httpClient})
			if err != nil {
				results <- result{err: err}
				return
			}
			defer conn.CloseNow()

			for {
				_, data, err := conn.Read(ctx)
				if err != nil {
					results <- result{err: err}
					return
				}
				var env protocol.Envelope
				if err := json.Unmarshal(data, &env); err != nil {
					results <- result{err: err}
					return
				}
				if env.Type == protocol.TypeSnapshot {
					results <- result{site: env.SiteID}
					<-releaseClients
					return
				}
			}
		}(i)
	}

	seen := make(map[uint64]struct{}, clientCount)
	var firstErr error
	var duplicate uint64
	for i := 0; i < clientCount; i++ {
		res := <-results
		if res.err != nil && firstErr == nil {
			firstErr = res.err
		}
		if _, ok := seen[res.site]; ok {
			duplicate = res.site
		}
		seen[res.site] = struct{}{}
	}
	close(releaseClients)
	clients.Wait()

	if firstErr != nil {
		t.Fatalf("join collaboration session: %v", firstErr)
	}
	if len(seen) != clientCount {
		t.Fatalf("%d concurrent sessions received only %d unique site IDs; duplicate site ID %d", clientCount, len(seen), duplicate)
	}
}
