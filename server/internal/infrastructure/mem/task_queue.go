package mem

import (
	"astrohop/internal/mission"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TaskQueue struct {
	queue       chan *mission.Task
	progressMap map[uuid.UUID]chan mission.TaskResult
	m           sync.RWMutex
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		queue:       make(chan *mission.Task),
		progressMap: make(map[uuid.UUID]chan mission.TaskResult),
	}
}

func (tq *TaskQueue) Queue(ctx context.Context) (chan *mission.Task, error) {
	return tq.queue, nil
}

func (tq *TaskQueue) Enqueue(ctx context.Context, task *mission.Task) error {
	ch := make(chan mission.TaskResult)
	_ = tq.SetTaskProgressChan(ctx, task.MissionID, ch)
	select {
	case <-time.After(time.Second * 10):
		_ = tq.SetTaskProgressChan(ctx, task.MissionID, nil)
		return errors.New("mission task TaskQueue: timeout on task enqueue")
	case tq.queue <- task:
		return nil
	}
}

func (tq *TaskQueue) SetTaskProgressChan(ctx context.Context, missionID uuid.UUID, progress chan mission.TaskResult) error {
	tq.m.Lock()
	defer tq.m.Unlock()
	if progress == nil {
		delete(tq.progressMap, missionID)
	}
	tq.progressMap[missionID] = progress
	return nil
}

func (tq *TaskQueue) GetTaskProgressChan(ctx context.Context, missionID uuid.UUID) (chan mission.TaskResult, error) {
	tq.m.RLock()
	defer tq.m.RUnlock()
	if ch, ok := tq.progressMap[missionID]; ok {
		return ch, nil
	}
	return nil, nil
}
