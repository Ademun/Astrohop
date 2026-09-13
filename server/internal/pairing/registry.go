package pairing

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

type SessionRegistry interface {
	Open(ctx context.Context, ttl time.Duration, conn *websocket.Conn) (string, error)
	Join(ctx context.Context, sessionID string, conn *websocket.Conn) error
	Push(ctx context.Context, to Role, sessionID string, payload []byte) error
	Close(ctx context.Context, sessionID string) error
}
