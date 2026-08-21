package crdt

import "encoding/json"

func (d *Doc) ApplyAll(ops []Op) error {
	for _, op := range ops {
		if err := d.Apply(op); err != nil {
			return err
		}
	}
	return nil
}

func (d *Doc) VisibleLen() int {
	return len([]rune(d.Text()))
}

type Stats struct {
	Atoms    int `json:"atoms"`
	Pending  int `json:"pending"`
	Visible  int `json:"visible"`
	Clock    uint64 `json:"clock"`
	Site     uint64 `json:"site"`
}

func (d *Doc) Stats() Stats {
	return Stats{
		Atoms: d.AtomCount(), Pending: d.PendingCount(),
		Visible: d.VisibleLen(), Clock: d.Clock(), Site: d.Site(),
	}
}

func MarshalOps(ops []Op) (string, error) {
	b, err := json.Marshal(ops)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func UnmarshalOps(raw string) ([]Op, error) {
	var ops []Op
	if err := json.Unmarshal([]byte(raw), &ops); err != nil {
		return nil, err
	}
	return ops, nil
}
