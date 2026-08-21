package crdt

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"
)

// Atom is a single code point in the RGA sequence.
type Atom struct {
	ID      ID     `json:"id"`
	After   ID     `json:"after"`
	Value   string `json:"value"`
	Deleted bool   `json:"deleted"`
}

type Doc struct {
	mu      sync.RWMutex
	site    uint64
	clock   uint64
	atoms   map[ID]*Atom
	seen    map[ID]struct{}
	pending []Op
}

func NewDoc(site uint64) *Doc {
	d := &Doc{
		site:  site,
		atoms: map[ID]*Atom{StartID: {ID: StartID, Value: ""}},
		seen:  map[ID]struct{}{StartID: {}},
	}
	return d
}

func (d *Doc) Site() uint64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.site
}

func (d *Doc) Clock() uint64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.clock
}

func (d *Doc) nextID() ID {
	d.clock++
	return ID{Site: d.site, Clock: d.clock}
}

func (d *Doc) bumpClock(c uint64) {
	if c > d.clock {
		d.clock = c
	}
}

// Apply is the pure mutation entry. Unknown causal parents are buffered
// until they arrive, so replicas can ingest ops in any order.
func (d *Doc) Apply(op Op) error {
	if err := op.Validate(); err != nil {
		return fmt.Errorf("invalid op: %w", err)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	queued, err := d.applyLocked(op)
	if err != nil {
		return err
	}
	if queued {
		d.pending = append(d.pending, op)
	}
	d.drainPending()
	return nil
}

func (d *Doc) applyLocked(op Op) (queued bool, err error) {
	if _, ok := d.seen[op.ID]; ok {
		return false, nil
	}
	switch op.Type {
	case OpInsert:
		if _, ok := d.atoms[op.After]; !ok {
			return true, nil
		}
		r, _ := utf8.DecodeRuneInString(op.Value)
		if r == utf8.RuneError && op.Value != string(utf8.RuneError) {
			return false, fmt.Errorf("insert value is not a valid rune")
		}
		d.atoms[op.ID] = &Atom{
			ID:    op.ID,
			After: op.After,
			Value: string(r),
		}
		d.seen[op.ID] = struct{}{}
		d.bumpClock(op.ID.Clock)
		return false, nil
	case OpDelete:
		target, ok := d.atoms[op.Target]
		if !ok {
			return true, nil
		}
		target.Deleted = true
		d.seen[op.ID] = struct{}{}
		d.bumpClock(op.ID.Clock)
		return false, nil
	default:
		return false, fmt.Errorf("unknown op type")
	}
}

func (d *Doc) drainPending() {
	if len(d.pending) == 0 {
		return
	}
	changed := true
	for changed {
		changed = false
		remain := d.pending[:0]
		for _, op := range d.pending {
			queued, err := d.applyLocked(op)
			if err != nil || queued {
				remain = append(remain, op)
				continue
			}
			changed = true
		}
		d.pending = remain
	}
}

// LocalInsert inserts s at the visible rune index and returns the generated ops.
func (d *Doc) LocalInsert(index int, s string) ([]Op, error) {
	if s == "" {
		return nil, nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	order := d.visibleIDsLocked()
	if index < 0 || index > len(order) {
		return nil, fmt.Errorf("insert index %d out of range 0..%d", index, len(order))
	}
	after := StartID
	if index > 0 {
		after = order[index-1]
	}
	ops := make([]Op, 0, utf8.RuneCountInString(s))
	for _, r := range s {
		id := d.nextID()
		op := Op{Type: OpInsert, ID: id, After: after, Value: string(r)}
		if queued, err := d.applyLocked(op); err != nil {
			return nil, err
		} else if queued {
			return nil, fmt.Errorf("local insert queued unexpectedly")
		}
		ops = append(ops, op)
		after = id
	}
	d.drainPending()
	return ops, nil
}

// LocalDelete tombstones `length` visible runes starting at index.
func (d *Doc) LocalDelete(index, length int) ([]Op, error) {
	if length <= 0 {
		return nil, nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	order := d.visibleIDsLocked()
	if index < 0 || index+length > len(order) {
		return nil, fmt.Errorf("delete range [%d,%d) out of range 0..%d", index, index+length, len(order))
	}
	ops := make([]Op, 0, length)
	for i := 0; i < length; i++ {
		id := d.nextID()
		op := Op{Type: OpDelete, ID: id, Target: order[index+i]}
		if queued, err := d.applyLocked(op); err != nil {
			return nil, err
		} else if queued {
			return nil, fmt.Errorf("local delete queued unexpectedly")
		}
		ops = append(ops, op)
	}
	d.drainPending()
	return ops, nil
}

func (d *Doc) visibleIDsLocked() []ID {
	var out []ID
	d.walk(StartID, func(a *Atom) {
		if !a.ID.IsStart() && !a.Deleted {
			out = append(out, a.ID)
		}
	})
	return out
}

func (d *Doc) childrenOf(parent ID) []*Atom {
	var kids []*Atom
	for _, a := range d.atoms {
		if a.After.Equal(parent) && !a.ID.Equal(parent) {
			kids = append(kids, a)
		}
	}
	sort.SliceStable(kids, func(i, j int) bool {
		return kids[i].ID.Greater(kids[j].ID)
	})
	return kids
}

func (d *Doc) walk(parent ID, fn func(*Atom)) {
	if a, ok := d.atoms[parent]; ok {
		fn(a)
	}
	for _, child := range d.childrenOf(parent) {
		d.walk(child.ID, fn)
	}
}

func (d *Doc) Text() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var b strings.Builder
	d.walk(StartID, func(a *Atom) {
		if a.ID.IsStart() || a.Deleted {
			return
		}
		b.WriteString(a.Value)
	})
	return b.String()
}

func (d *Doc) PendingCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.pending)
}

func (d *Doc) AtomCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.atoms)
}

func (d *Doc) Snapshot() Snapshot {
	d.mu.RLock()
	defer d.mu.RUnlock()
	atoms := make([]Atom, 0, len(d.atoms))
	for _, a := range d.atoms {
		if a.ID.IsStart() {
			continue
		}
		atoms = append(atoms, *a)
	}
	sort.Slice(atoms, func(i, j int) bool { return atoms[i].ID.Less(atoms[j].ID) })
	return Snapshot{Site: d.site, Clock: d.clock, Atoms: atoms}
}

func (d *Doc) LoadSnapshot(s Snapshot) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.site = s.Site
	d.clock = s.Clock
	d.atoms = map[ID]*Atom{StartID: {ID: StartID}}
	d.seen = map[ID]struct{}{StartID: {}}
	d.pending = nil
	for i := range s.Atoms {
		a := &s.Atoms[i]
		d.atoms[a.ID] = a
		d.seen[a.ID] = struct{}{}
	}
}

type Snapshot struct {
	Site  uint64 `json:"site"`
	Clock uint64 `json:"clock"`
	Atoms []Atom `json:"atoms"`
}

func (s Snapshot) Marshal() (string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ParseSnapshot(raw string) (Snapshot, error) {
	var s Snapshot
	if strings.TrimSpace(raw) == "" {
		return s, nil
	}
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return s, err
	}
	return s, nil
}

func Clone(src *Doc, site uint64) *Doc {
	snap := src.Snapshot()
	dst := NewDoc(site)
	dst.LoadSnapshot(snap)
	dst.mu.Lock()
	dst.site = site
	if snap.Clock > dst.clock {
		dst.clock = snap.Clock
	}
	dst.mu.Unlock()
	return dst
}
