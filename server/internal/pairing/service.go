package pairing

import (
	"astrohop/pkg/apperr"
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type SessionRegistry interface {
	Open(ctx context.Context, ttl time.Duration, conn *websocket.Conn) (string, error)
	Join(ctx context.Context, sessionID string, conn *websocket.Conn) error
	Push(ctx context.Context, to Role, sessionID string, payload []byte) error
	Next(ctx context.Context, from Role, sessionID string) ([]byte, error)
	Close(ctx context.Context, sessionID string) error
}

type Service struct {
	registry SessionRegistry
}

func NewService(registry SessionRegistry) *Service {
	return &Service{
		registry: registry,
	}
}

func (s *Service) StartPairing(ctx context.Context, conn *websocket.Conn) error {
	sessionID, err := s.registry.Open(ctx, 2*time.Minute, conn)
	if err != nil {
		return regErr(err)
	}
	if err := s.registry.Push(ctx, RoleSource, sessionID, []byte(sessionID)); err != nil {
		return regErr(err)
	}
	return nil
}

func (s *Service) AttachTarget(ctx context.Context, conn *websocket.Conn, sessionID string, encryptionKey string) error {
	if strings.TrimSpace(encryptionKey) == "" {
		return apperr.Validation(ErrKey, "encryption key is empty", nil)
	}

	if err := s.registry.Join(ctx, sessionID, conn); err != nil {
		return regErr(err)
	}
	if err := s.registry.Push(ctx, RoleSource, sessionID, []byte(encryptionKey)); err != nil {
		return regErr(err)
	}

	ctxTm, cancelTm := context.WithTimeout(ctx, 10*time.Second)
	defer cancelTm()

	accountKey, err := s.registry.Next(ctxTm, RoleSource, sessionID)
	if err != nil {
		return regErr(err)
	}

	if err := s.registry.Push(ctx, RoleTarget, sessionID, accountKey); err != nil {
		return regErr(err)
	}

	return regErr(s.registry.Close(ctx, sessionID))
}

func regErr(err error) error {
	var netErr net.Error

	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrNotFound):
		return apperr.NotFound(ErrSession, "session not found", err)
	case errors.Is(err, ErrAlreadyJoined):
		return apperr.Conflict(ErrSession, "session was already joined", err)
	case errors.As(err, &netErr) && netErr.Timeout():
		return apperr.Timeout(ErrSession, "session timed out waiting for peer", err)
	default:
		return apperr.Internal(ErrSession, "unknown session registry error", err)
	}
}
