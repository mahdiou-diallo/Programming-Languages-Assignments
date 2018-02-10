package chatty

/* Library Code */

// Enum
type Purpose int

const (
	CONNECT Purpose = 1 + iota
	MSG
	LIST
	ERROR
	DISCONNECT
)

type ChattyMsg struct {
	Username string
	Body     string
	Action   Purpose
}
