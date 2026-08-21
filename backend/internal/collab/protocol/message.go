package protocol

import "gowiki/internal/collab/crdt"

const (
	TypeAuth     = "auth"
	TypeJoin     = "join"
	TypeOp       = "op"
	TypePresence = "presence"
	TypeLock     = "lock"
	TypeSync     = "sync"
	TypeSnapshot = "snapshot"
	TypeError    = "error"
	TypePong     = "pong"
	TypePing     = "ping"
)

type Envelope struct {
	Type       string      `json:"type"`
	Token      string      `json:"token,omitempty"`
	DocumentID string      `json:"documentId,omitempty"`
	Op         *crdt.Op    `json:"op,omitempty"`
	Cursor     int         `json:"cursor,omitempty"`
	Color      string      `json:"color,omitempty"`
	Paragraph  string      `json:"paragraphId,omitempty"`
	Action     string      `json:"action,omitempty"`
	SinceClock uint64      `json:"sinceClock,omitempty"`
	Text       string      `json:"text,omitempty"`
	Clock      uint64      `json:"clock,omitempty"`
	SiteID     uint64      `json:"siteId,omitempty"`
	Users      []Presence  `json:"users,omitempty"`
	Holder     string      `json:"holder,omitempty"`
	Until      string      `json:"until,omitempty"`
	Code       string      `json:"code,omitempty"`
	Message    string      `json:"message,omitempty"`
	Locks      []LockState `json:"locks,omitempty"`
	Atoms      []crdt.Atom `json:"atoms,omitempty"`
}

type Presence struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Cursor int    `json:"cursor"`
}

type LockState struct {
	ParagraphID string `json:"paragraphId"`
	Holder      string `json:"holder"`
	HolderName  string `json:"holderName"`
	Until       string `json:"until"`
}
