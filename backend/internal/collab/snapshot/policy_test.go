package snapshot

import (
	"testing"
	"time"
)

func TestEvaluateTriggers(t *testing.T) {
	if Evaluate(time.Second, 10, time.Minute, 0, false).PersistOps {
		t.Fatal("no pending")
	}
	d := Evaluate(time.Second, 500, time.Minute, 3, false)
	if !d.AutoL2 || d.Reason != "chars" {
		t.Fatalf("%+v", d)
	}
	d = Evaluate(30*time.Second, 1, time.Minute, 1, false)
	if d.Reason != "idle" {
		t.Fatalf("%+v", d)
	}
	d = Evaluate(time.Second, 1, 10*time.Minute, 1, false)
	if d.Reason != "interval" {
		t.Fatalf("%+v", d)
	}
	if !Evaluate(0, 0, 0, 1, true).AutoL2 {
		t.Fatal("force")
	}
}

func TestRetainAndTimeouts(t *testing.T) {
	if !RetainL2(50) || RetainL2(51) {
		t.Fatal("retain")
	}
	if OpRetention() != 24*time.Hour {
		t.Fatal("ops")
	}
	if LockHeartbeat() != 60*time.Second || LockTimeout() != 180*time.Second {
		t.Fatal("lock")
	}
}
