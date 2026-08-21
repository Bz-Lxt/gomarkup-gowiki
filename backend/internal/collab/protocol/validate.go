package protocol

import (
	"fmt"
	"strings"

	"gowiki/internal/collab/crdt"
)

func ValidateEnvelope(env Envelope) error {
	switch env.Type {
	case TypeAuth:
		if strings.TrimSpace(env.Token) == "" {
			return fmt.Errorf("auth token required")
		}
	case TypeJoin:
		if strings.TrimSpace(env.DocumentID) == "" {
			return fmt.Errorf("documentId required")
		}
	case TypeOp:
		if env.Op == nil {
			return fmt.Errorf("op required")
		}
		return env.Op.Validate()
	case TypeLock:
		if strings.TrimSpace(env.Paragraph) == "" {
			return fmt.Errorf("paragraphId required")
		}
		switch env.Action {
		case "acquire", "heartbeat", "release":
		default:
			return fmt.Errorf("unknown lock action")
		}
	case TypePresence, TypePing, TypePong, TypeSync, TypeSnapshot, TypeError:
		return nil
	case "":
		return fmt.Errorf("empty message type")
	default:
		return fmt.Errorf("unknown type %s", env.Type)
	}
	return nil
}

func NewError(code, msg string) Envelope {
	return Envelope{Type: TypeError, Code: code, Message: msg}
}

func NewOp(op crdt.Op, site uint64) Envelope {
	return Envelope{Type: TypeOp, Op: &op, SiteID: site}
}
