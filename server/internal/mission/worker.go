package mission

import (
	"context"

	"astrohop/internal/mission/planner"
	"astrohop/pkg/apperr"

	"github.com/google/uuid"
)

func (s *Service) Start(ctx context.Context) {
	for range 5 {
		go s.missionWorker(ctx)
	}
}

func (s *Service) missionWorker(ctx context.Context) {
	for {
		task, err := s.taskQueue.Dequeue(ctx)
		if err != nil {
			s.log.Error("mission worker: dequeue failed", "cause", err.Error())
			continue
		}
		s.runTask(ctx, task)
	}
}

func (s *Service) runTask(ctx context.Context, task *Task) {
	defer s.inflight.Delete(task.MissionID)
	defer func() {
		if err := s.progressHub.Close(ctx, task.MissionID); err != nil {
			s.log.Error("mission worker: close progress hub failed",
				"mission", task.MissionID.String(), "cause", err.Error())
		}
	}()
	defer func() {
		if r := recover(); r != nil {
			s.log.Error("mission worker: panic", "mission", task.MissionID.String(), "panic", r)
		}
	}()

	s.publish(ctx, task.MissionID, TaskResult{Progress: taskBuildingRoute})

	input, err := s.buildPlannerInput(ctx, task.Data)
	if err != nil {
		s.publish(ctx, task.MissionID, TaskResult{Progress: taskFailed, Error: err})
		return
	}

	output := planner.Build(input)
	mapData := &Map{
		MoonPosition: output.MoonPosition,
		Positions:    output.Positions,
		Tour:         output.Tour,
	}

	if err := s.missionRepo.UpdateMissionMap(ctx, task.DataVersion, mapData, task.MissionID); err != nil {
		s.publish(ctx, task.MissionID, TaskResult{
			Progress: taskFailed,
			Error:    apperr.Internal(ErrUpdateMission, "failed to persist map data", err),
		})
		return
	}

	s.publish(ctx, task.MissionID, TaskResult{Progress: taskDone, Payload: mapData})
}

func (s *Service) buildPlannerInput(ctx context.Context, data *Data) (*planner.Input, error) {
	oids := make([]int64, len(data.Objectives))
	for i, o := range data.Objectives {
		oids[i] = o.OID
	}

	stellarData, err := s.searchSvc.GetObjectsStellarData(ctx, oids)
	if err != nil {
		return nil, err
	}

	objectives := make([]planner.Objective, len(stellarData))
	for i, d := range stellarData {
		objectives[i] = planner.Objective{OID: oids[i], Stellar: d}
	}

	return &planner.Input{
		Location:   data.Location,
		Time:       data.Time,
		Objectives: objectives,
	}, nil
}

func (s *Service) enqueue(ctx context.Context, t *Task) error {
	if _, loaded := s.inflight.LoadOrStore(t.MissionID, struct{}{}); loaded {
		return nil
	}
	if err := s.taskQueue.Enqueue(ctx, t); err != nil {
		s.inflight.Delete(t.MissionID)
		return err
	}
	return nil
}

func (s *Service) publish(ctx context.Context, id uuid.UUID, r TaskResult) {
	if err := s.progressHub.Pub(ctx, id, r); err != nil {
		s.log.Error("mission worker: publish failed",
			"mission", id.String(), "cause", err.Error())
	}
}
