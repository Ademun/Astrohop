package mem

import (
	"astrohop/internal/mission"
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

const publishTimeout = 5 * time.Second

type ProgressHub struct {
	mu     sync.Mutex
	topics map[uuid.UUID][]*memSubscriber
}

func NewProgressHub() *ProgressHub {
	return &ProgressHub{
		topics: make(map[uuid.UUID][]*memSubscriber),
	}
}

func (h *ProgressHub) Pub(ctx context.Context, missionID uuid.UUID, progress mission.TaskResult) error {
	h.mu.Lock()
	subs := h.topics[missionID]
	snapshot := make([]*memSubscriber, len(subs))
	copy(snapshot, subs)
	h.mu.Unlock()

	for _, s := range snapshot {
		s.deliver(ctx, progress, publishTimeout)
	}
	return nil
}

func (h *ProgressHub) Sub(_ context.Context, missionID uuid.UUID) (<-chan mission.TaskResult, func(), error) {
	sub := newMemSubscriber()

	h.mu.Lock()
	h.topics[missionID] = append(h.topics[missionID], sub)
	h.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			h.mu.Lock()
			subs := h.topics[missionID]
			for i, s := range subs {
				if s == sub {
					h.topics[missionID] = append(subs[:i], subs[i+1:]...)
					break
				}
			}
			h.mu.Unlock()
			sub.close()
		})
	}

	return sub.ch, cancel, nil
}

func (h *ProgressHub) Close(_ context.Context, missionID uuid.UUID) error {
	h.mu.Lock()
	subs := h.topics[missionID]
	delete(h.topics, missionID)
	h.mu.Unlock()

	for _, s := range subs {
		s.close()
	}
	return nil
}

type memSubscriber struct {
	mu      sync.Mutex
	ch      chan mission.TaskResult
	done    chan struct{}
	closed  bool
	senders sync.WaitGroup
	once    sync.Once
}

func newMemSubscriber() *memSubscriber {
	return &memSubscriber{
		ch:   make(chan mission.TaskResult),
		done: make(chan struct{}),
	}
}

func (s *memSubscriber) deliver(ctx context.Context, r mission.TaskResult, timeout time.Duration) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.senders.Add(1)
	s.mu.Unlock()
	defer s.senders.Done()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case s.ch <- r:
	case <-s.done:
	case <-ctx.Done():
	case <-timer.C:
	}
}

func (s *memSubscriber) close() {
	s.once.Do(func() {
		s.mu.Lock()
		s.closed = true
		close(s.done)
		s.mu.Unlock()

		s.senders.Wait()
		close(s.ch)
	})
}
