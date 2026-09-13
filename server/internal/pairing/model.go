package pairing

import "time"

type SessionStatus string

const (
	StatusAwaitingTarget  SessionStatus = "AWAITING_TARGET"
	StatusAwaitingAccount SessionStatus = "AWAITING_ACCOUNT"
)

type Session struct {
	SessionID string        `json:"session_id"`
	Status    SessionStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	ExpiresAt time.Time     `json:"expires_at"`
}
