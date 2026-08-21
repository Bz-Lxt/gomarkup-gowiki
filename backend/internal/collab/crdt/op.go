package crdt

import "fmt"

type OpType string

const (
	OpInsert OpType = "insert"
	OpDelete OpType = "delete"
)

// Op is a single causal mutation. Insert.Value is one Unicode code point
// encoded as a string (may be a multi-byte UTF-8 sequence). Multi-rune
// local edits are expanded into a chain of Ops by Doc.LocalInsert.
type Op struct {
	Type   OpType `json:"type"`
	ID     ID     `json:"id"`
	After  ID     `json:"after"`
	Target ID     `json:"target"`
	Value  string `json:"value,omitempty"`
}

func (o Op) Key() ID {
	if o.Type == OpDelete {
		return o.ID
	}
	return o.ID
}

func (o Op) Validate() error {
	switch o.Type {
	case OpInsert:
		if o.ID.IsStart() {
			return fmt.Errorf("insert id cannot be start sentinel")
		}
		if o.Value == "" {
			return fmt.Errorf("insert value is empty")
		}
		return nil
	case OpDelete:
		if o.Target.IsStart() {
			return fmt.Errorf("cannot delete start sentinel")
		}
		if o.ID.IsStart() {
			return fmt.Errorf("delete op id cannot be start sentinel")
		}
		return nil
	default:
		return fmt.Errorf("unknown op type %q", o.Type)
	}
}
