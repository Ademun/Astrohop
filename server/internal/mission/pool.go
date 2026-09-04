package mission

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type pool struct {
	Queue       chan missionTask
	progressMap map[uuid.UUID]chan taskResult
	m           sync.RWMutex
}

func newPool() *pool {
	return &pool{
		Queue:       make(chan missionTask),
		progressMap: make(map[uuid.UUID]chan taskResult),
	}
}

func (p *pool) enqueueTask(task missionTask) error {
	ch := make(chan taskResult)
	p.setTaskProgressChan(task.MissionID, ch)
	select {
	case <-time.After(time.Second * 10):
		p.setTaskProgressChan(task.MissionID, nil)
		return errors.New("mission task pool: timeout on task enqueue")
	case p.Queue <- task:
		return nil
	}
}

func (p *pool) setTaskProgressChan(missionID uuid.UUID, progress chan taskResult) {
	p.m.Lock()
	defer p.m.Unlock()
	if progress == nil {
		delete(p.progressMap, missionID)
	}
	p.progressMap[missionID] = progress
}

func (p *pool) getTaskProgressChan(missionID uuid.UUID) chan taskResult {
	p.m.RLock()
	defer p.m.RUnlock()
	if ch, ok := p.progressMap[missionID]; ok {
		return ch
	}
	return nil
}
