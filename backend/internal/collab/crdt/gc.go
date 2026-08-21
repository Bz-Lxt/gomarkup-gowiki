package crdt

// GC removes tombstones that are not referenced as an After parent.
// It is safe: remaining atoms still reconstruct the same visible text.
func (d *Doc) GC() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	referenced := map[ID]struct{}{StartID: {}}
	for _, a := range d.atoms {
		referenced[a.After] = struct{}{}
	}
	removed := 0
	for id, a := range d.atoms {
		if id.IsStart() || !a.Deleted {
			continue
		}
		if _, used := referenced[id]; used {
			continue
		}
		delete(d.atoms, id)
		removed++
	}
	return removed
}
