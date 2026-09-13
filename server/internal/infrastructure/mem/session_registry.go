package mem

import (
	"astrohop/internal/pairing"
	"astrohop/pkg/apperr"
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type SessionRegistry struct {
	sessions map[string]*pairing.Session
	lock     sync.RWMutex
}

func New() *SessionRegistry {
	return &SessionRegistry{
		sessions: make(map[string]*pairing.Session),
	}
}

func (s *SessionRegistry) Open(ctx context.Context, ttl time.Duration, conn *websocket.Conn) (string, error) {
	id := uuid.New().String()
	session := &pairing.Session{
		SourceConn: conn,
		Status:     pairing.StatusAwaitingTarget,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(ttl),
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.sessions[id] = session
	time.AfterFunc(ttl, func() {
		s.cleanup(id)
	})
	return id, s.Push(ctx, pairing.RoleSource, id, []byte(id))
}

func (s *SessionRegistry) Join(ctx context.Context, sessionID string, conn *websocket.Conn) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	session, ok := s.sessions[sessionID]
	if !ok {
		return apperr.New(http.StatusGone, "session expired", nil)
	}
	session.TargetConn = conn
	return nil
}

func (s *SessionRegistry) Push(ctx context.Context, to pairing.Role, sessionID string, payload []byte) error {
	s.lock.RLock()
	session, ok := s.sessions[sessionID]
	if !ok {
		return apperr.New(http.StatusGone, "session expired", nil)
	}
	s.lock.RUnlock()
	switch to {
	case pairing.RoleSource:
		return session.SourceConn.WriteMessage(websocket.BinaryMessage, payload)
	case pairing.RoleTarget:
		return session.TargetConn.WriteMessage(websocket.BinaryMessage, payload)
	}
	return nil
}

func (s *SessionRegistry) Close(ctx context.Context, sessionID string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	session, ok := s.sessions[sessionID]
	if !ok {
		return apperr.New(http.StatusGone, "session expired", nil)
	}
	if err := session.SourceConn.Close(); err != nil {
		return err
	}
	if err := session.TargetConn.Close(); err != nil {
		return err
	}
	delete(s.sessions, sessionID)
	return nil
}

func (s *SessionRegistry) cleanup(sessionID string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.sessions, sessionID)
}
