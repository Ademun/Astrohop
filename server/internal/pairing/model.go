package pairing

import (
	"time"

	"github.com/gorilla/websocket"
)

type SessionStatus string

const (
	StatusAwaitingTarget  SessionStatus = "AWAITING_TARGET"
	StatusAwaitingAccount SessionStatus = "AWAITING_ACCOUNT"
)

type Role string

const (
	RoleSource Role = "SOURCE"
	RoleTarget Role = "TARGET"
)

type Session struct {
	SourceConn *websocket.Conn
	TargetConn *websocket.Conn
	Status     SessionStatus `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	ExpiresAt  time.Time     `json:"expires_at"`
}
