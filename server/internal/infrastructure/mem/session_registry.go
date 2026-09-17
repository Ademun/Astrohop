package mem

import (
	"astrohop/internal/pairing"
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type entry struct {
	session *pairing.Session
	timer   *time.Timer
	closed  bool
}

type SessionRegistry struct {
	log      *slog.Logger
	sessions map[string]*entry
	lock     sync.RWMutex
}

func NewSessionRegistry(log *slog.Logger) *SessionRegistry {
	return &SessionRegistry{
		log:      log,
		sessions: make(map[string]*entry),
	}
}

func (s *SessionRegistry) Open(_ context.Context, ttl time.Duration, conn *websocket.Conn) (string, error) {
	id := uuid.New().String()
	e := &entry{
		session: &pairing.Session{
			SourceConn: conn,
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(ttl),
		},
	}

	s.lock.Lock()
	s.sessions[id] = e
	e.timer = time.AfterFunc(ttl, func() {
		if err := s.Close(context.Background(), id); err != nil {
			s.log.Error("failed to expire pairing session", "session", id, "cause", err.Error())
		}
	})
	s.lock.Unlock()

	return id, nil
}

func (s *SessionRegistry) Join(_ context.Context, sessionID string, conn *websocket.Conn) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	e, ok := s.sessions[sessionID]
	if !ok {
		return pairing.ErrNotFound
	}
	if e.session.TargetConn != nil {
		return pairing.ErrAlreadyJoined
	}
	e.session.TargetConn = conn
	return nil
}

func (s *SessionRegistry) Push(ctx context.Context, to pairing.Role, sessionID string, payload []byte) error {
	c, err := s.conn(sessionID, to)
	if err != nil {
		return err
	}
	if deadline, ok := ctx.Deadline(); ok {
		if err := c.SetWriteDeadline(deadline); err != nil {
			return err
		}
	}
	return c.WriteMessage(websocket.BinaryMessage, payload)
}

func (s *SessionRegistry) Next(ctx context.Context, from pairing.Role, sessionID string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	c, err := s.conn(sessionID, from)
	if err != nil {
		return nil, err
	}

	if deadline, ok := ctx.Deadline(); ok {
		if err := c.SetReadDeadline(deadline); err != nil {
			return nil, err
		}
	}

	_, msg, err := c.ReadMessage()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *SessionRegistry) Close(_ context.Context, sessionID string) error {
	s.lock.Lock()
	e, ok := s.sessions[sessionID]
	if !ok {
		s.lock.Unlock()
		return pairing.ErrNotFound
	}
	if e.closed {
		s.lock.Unlock()
		return nil
	}
	e.closed = true
	e.timer.Stop()
	delete(s.sessions, sessionID)
	s.lock.Unlock()

	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	var errs []error

	if c := e.session.SourceConn; c != nil {
		if err := c.WriteMessage(websocket.CloseMessage, closeMsg); err != nil {
			errs = append(errs, err)
		}
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if c := e.session.TargetConn; c != nil {
		if err := c.WriteMessage(websocket.CloseMessage, closeMsg); err != nil {
			errs = append(errs, err)
		}
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		s.log.Error("errors while closing pairing session", "session", sessionID, "cause", errors.Join(errs...).Error())
	}
	return nil
}

func (s *SessionRegistry) conn(sessionID string, role pairing.Role) (*websocket.Conn, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	e, ok := s.sessions[sessionID]
	if !ok {
		return nil, pairing.ErrNotFound
	}
	switch role {
	case pairing.RoleSource:
		if e.session.SourceConn == nil {
			return nil, pairing.ErrNotFound
		}
		return e.session.SourceConn, nil
	case pairing.RoleTarget:
		if e.session.TargetConn == nil {
			return nil, pairing.ErrNotFound
		}
		return e.session.TargetConn, nil
	default:
		return nil, pairing.ErrUnsupportedRole
	}
}
