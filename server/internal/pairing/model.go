package pairing

import (
	"errors"
	"time"

	"github.com/gorilla/websocket"
)

type Role string

const (
	RoleSource Role = "SOURCE"
	RoleTarget Role = "TARGET"
)

type SessionErr error

var (
	ErrExpired         SessionErr = errors.New("expired")
	ErrNotFound        SessionErr = errors.New("not found")
	ErrAlreadyJoined   SessionErr = errors.New("already joined")
	ErrUnsupportedRole SessionErr = errors.New("unsupported role")
)

type Session struct {
	SourceConn *websocket.Conn
	TargetConn *websocket.Conn
	CreatedAt  time.Time
	ExpiresAt  time.Time
}
