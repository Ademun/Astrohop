package mission

import (
	"astrohop/internal/mission/planner"
	"astrohop/internal/search"
	"astrohop/pkg/apperr"
	"context"
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

type Repo interface {
	CreateMission(ctx context.Context, data *Data, accountID int64) (uuid.UUID, error)
	GetMission(ctx context.Context, id uuid.UUID) (*Mission, error)
	GetAccountMissions(ctx context.Context, accountID int64) ([]Mission, error)
	UpdateMissionData(ctx context.Context, missionID uuid.UUID, data *Data) error
	UpdateMissionMapData(ctx context.Context, missionID uuid.UUID, data *MapData) error
	UpdateMissionVisibility(ctx context.Context, missionID uuid.UUID, isPublic bool) error
	DeleteMission(ctx context.Context, missionID uuid.UUID) error
}

type TaskQueue interface {
	Enqueue(ctx context.Context, task *Task) error
	Dequeue(ctx context.Context) (*Task, error)
}

type ProgressHub interface {
	Pub(ctx context.Context, missionID uuid.UUID, progress TaskResult) error
	Sub(ctx context.Context, missionID uuid.UUID) (<-chan TaskResult, func(), error)
	Close(ctx context.Context, missionID uuid.UUID) error
}

type Service struct {
	missionRepo Repo
	taskQueue   TaskQueue
	progressHub ProgressHub
	searchSvc   *search.Service
	inflight    sync.Map
	log         *slog.Logger
}

func NewService(
	missionRepo Repo,
	taskQueue TaskQueue,
	progressHub ProgressHub,
	searchSvc *search.Service,
	log *slog.Logger,
) *Service {
	return &Service{
		searchSvc:   searchSvc,
		missionRepo: missionRepo,
		taskQueue:   taskQueue,
		progressHub: progressHub,
		log:         log,
	}
}

func (s *Service) Start(ctx context.Context) {
	for range 5 {
		go s.missionWorker(ctx)
	}
}

func (s *Service) CreateMission(ctx context.Context, data *Data, accountID int64) (uuid.UUID, error) {
	missionID, err := s.missionRepo.CreateMission(ctx, data, accountID)
	if err != nil {
		return uuid.Nil, apperr.Internal(ErrCreateMission, "failed to create new mission", err)
	}
	task := &Task{
		MissionID: missionID,
		Data:      data,
	}
	if err := s.taskQueue.Enqueue(ctx, task); err != nil {
		return uuid.Nil, apperr.Internal(ErrTaskQueue, "failed to enqueue task", err)
	}
	return missionID, nil
}

func (s *Service) UpdateMissionData(ctx context.Context, missionID uuid.UUID, data *Data) error {
	if err := s.missionRepo.UpdateMissionData(ctx, missionID, data); err != nil {
		return apperr.Internal(ErrUpdateMission, "failed to update mission data", err)
	}
	task := &Task{
		MissionID: missionID,
		Data:      data,
	}
	if err := s.taskQueue.Enqueue(ctx, task); err != nil {
		return apperr.Internal(ErrTaskQueue, "failed to enqueue task", err)
	}
	return nil
}

func (s *Service) UpdateMissionVisibility(ctx context.Context, missionID uuid.UUID, isPublic bool) error {
	if err := s.missionRepo.UpdateMissionVisibility(ctx, missionID, isPublic); err != nil {
		return apperr.Internal(ErrUpdateMission, "failed to update mission visibility", err)
	}
	return nil
}

func (s *Service) DeleteMission(ctx context.Context, missionID uuid.UUID) error {
	if err := s.missionRepo.DeleteMission(ctx, missionID); err != nil {
		return apperr.Internal(ErrDeleteMission, "failed to delete mission", err)
	}
	return nil
}

func (s *Service) GetMission(ctx context.Context, id uuid.UUID) (*Mission, error) {
	mission, err := s.missionRepo.GetMission(ctx, id)
	if err != nil {
		return nil, apperr.Internal(ErrGetMission, "failed to get mission", err)
	}
	return mission, nil
}

func (s *Service) GetAccountMissions(ctx context.Context, accountID int64) ([]Mission, error) {
	missions, err := s.missionRepo.GetAccountMissions(ctx, accountID)
	if err != nil {
		return nil, apperr.Internal(ErrGetMission, "failed to get account missions", err)
	}
	return missions, nil
}

func (s *Service) GetMissionStream(ctx context.Context, id uuid.UUID) (<-chan TaskResult, func(), error) {
	mission, err := s.missionRepo.GetMission(ctx, id)
	if err != nil {
		return nil, nil, apperr.Internal(ErrGetMission, "failed to load mission", err)
	}

	ch, cancel, err := s.progressHub.Sub(ctx, id)
	if err != nil {
		return nil, nil, apperr.Internal(ErrProgressHub, "failed to subscribe", err)
	}

	if mission.MapData != nil {
		if err := s.progressHub.Pub(ctx, id, TaskResult{Progress: taskDone, Payload: mission.MapData}); err != nil {
			cancel()
			return nil, nil, apperr.Internal(ErrProgressHub, "failed to publish", err)
		}
		return ch, cancel, nil
	}

	if err := s.enqueue(ctx, &Task{MissionID: id, Data: mission.Data}); err != nil {
		cancel()
		return nil, nil, apperr.Internal(ErrTaskQueue, "failed to enqueue task", err)
	}
	return ch, cancel, nil
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
	mapData := &MapData{
		MoonPosition: output.MoonPosition,
		Positions:    output.Positions,
		Tour:         output.Tour,
	}

	if err := s.missionRepo.UpdateMissionMapData(ctx, task.MissionID, mapData); err != nil {
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
