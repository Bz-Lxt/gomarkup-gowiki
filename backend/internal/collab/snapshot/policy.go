package snapshot

import "time"

// Decision encodes the C-6 three-layer snapshot policy.
type Decision struct {
	PersistOps bool
	AutoL2     bool
	Reason     string
}

func Evaluate(idle time.Duration, dirtyRunes int, sinceLast time.Duration, pending int, force bool) Decision {
	if pending == 0 && !force {
		return Decision{}
	}
	d := Decision{PersistOps: pending > 0}
	if force {
		d.AutoL2 = true
		d.Reason = "force"
		return d
	}
	if dirtyRunes >= 500 {
		d.AutoL2 = true
		d.Reason = "chars"
		return d
	}
	if sinceLast >= 10*time.Minute {
		d.AutoL2 = true
		d.Reason = "interval"
		return d
	}
	if idle >= 30*time.Second {
		d.AutoL2 = true
		d.Reason = "idle"
		return d
	}
	return d
}

func RetainL2(count int) bool {
	return count <= 50
}

func OpRetention() time.Duration {
	return 24 * time.Hour
}

func LockHeartbeat() time.Duration { return 60 * time.Second }
func LockTimeout() time.Duration   { return 180 * time.Second }
