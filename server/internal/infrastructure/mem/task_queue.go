package mem

import (
	"astrohop/internal/mission"
	"context"
)

type TaskQueue struct {
	tasks chan *mission.Task
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{tasks: make(chan *mission.Task)}
}

func (q *TaskQueue) Enqueue(ctx context.Context, task *mission.Task) error {
	select {
	case q.tasks <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *TaskQueue) Dequeue(ctx context.Context) (*mission.Task, error) {
	select {
	case task := <-q.tasks:
		return task, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
