package protocol

import (
	"testing"

	"gowiki/internal/collab/crdt"
)

func TestValidateEnvelope(t *testing.T) {
	if err := ValidateEnvelope(Envelope{Type: TypeOp}); err == nil {
		t.Fatal("missing op")
	}
	op := crdt.Op{Type: crdt.OpInsert, ID: crdt.ID{Site: 1, Clock: 1}, After: crdt.StartID, Value: "a"}
	if err := ValidateEnvelope(Envelope{Type: TypeOp, Op: &op}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateEnvelope(Envelope{Type: TypeLock, Paragraph: "p1", Action: "acquire"}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateEnvelope(Envelope{Type: TypeLock, Paragraph: "p1", Action: "nope"}); err == nil {
		t.Fatal("bad action")
	}
	if err := ValidateEnvelope(Envelope{Type: "zzz"}); err == nil {
		t.Fatal("unknown")
	}
}
