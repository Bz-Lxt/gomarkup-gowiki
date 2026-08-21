package lock

import (
	"testing"
	"time"
)

func TestLockAcquireAndTimeout(t *testing.T) {
	s := New(40 * time.Millisecond)
	_, ok := s.Acquire("d1", "p1", "u1", "A")
	if !ok {
		t.Fatal("first acquire")
	}
	_, ok = s.Acquire("d1", "p1", "u2", "B")
	if ok {
		t.Fatal("should be locked")
	}
	time.Sleep(50 * time.Millisecond)
	_, ok = s.Acquire("d1", "p1", "u2", "B")
	if !ok {
		t.Fatal("timeout should release")
	}
}

func TestHeartbeatAndRelease(t *testing.T) {
	s := New(time.Second)
	_, _ = s.Acquire("d", "p", "u", "N")
	if _, ok := s.Heartbeat("d", "p", "u"); !ok {
		t.Fatal("heartbeat")
	}
	s.Release("d", "p", "u")
	if len(s.List("d")) != 0 {
		t.Fatal("expected empty")
	}
}
