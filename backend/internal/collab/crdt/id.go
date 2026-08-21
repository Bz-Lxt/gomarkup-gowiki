package crdt

import "fmt"

// ID uniquely identifies an atom in an RGA sequence.
type ID struct {
	Site  uint64 `json:"site"`
	Clock uint64 `json:"clock"`
}

// StartID is the immutable sentinel at the head of every document.
var StartID = ID{Site: 0, Clock: 0}

func (a ID) Equal(b ID) bool { return a.Site == b.Site && a.Clock == b.Clock }

func (a ID) Less(b ID) bool {
	if a.Clock != b.Clock {
		return a.Clock < b.Clock
	}
	return a.Site < b.Site
}

func (a ID) Greater(b ID) bool { return b.Less(a) }

func (a ID) String() string { return fmt.Sprintf("%d:%d", a.Site, a.Clock) }

func (a ID) IsStart() bool { return a.Equal(StartID) }
