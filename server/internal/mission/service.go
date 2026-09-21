package mission

import (
	"astrohop/internal/search"
	"astrohop/pkg/apperr"
	"context"
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

type Repo interface {
	CreateMission(ctx context.Context, accountID int64, data *Data) (uuid.UUID, error)
	GetMission(ctx context.Context, missionID uuid.UUID) (*Mission, error)
	GetAccountMissions(ctx context.Context, accountID int64) ([]Mission, error)
	GetMissionOwner(ctx context.Context, missionID uuid.UUID) (int64, error)
	UpdateMissionInformation(ctx context.Context, info *Information, missionID uuid.UUID) error
	UpdateMissionData(ctx context.Context, data *Data, missionID uuid.UUID) (int, error)
	UpdateMissionMap(ctx context.Context, sourceDataVersion int, m *Map, missionID uuid.UUID) error
	SetMissionMapOverrides(ctx context.Context, overrides *Overrides, missionID uuid.UUID) error
	SetMissionVisibility(ctx context.Context, isPublic bool, missionID uuid.UUID) error
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

func (s *Service) CreateMission(ctx context.Context, data *Data, accountID int64) (uuid.UUID, error) {
	missionID, err := s.missionRepo.CreateMission(ctx, accountID, data)
	if err != nil {
		return uuid.Nil, apperr.Internal(ErrCreateMission, "failed to create new mission", err)
	}
	task := &Task{
		MissionID:   missionID,
		DataVersion: 0,
		Data:        data,
	}
	if err := s.enqueue(ctx, task); err != nil {
		return uuid.Nil, apperr.Internal(ErrTaskQueue, "failed to enqueue task", err)
	}
	return missionID, nil
}

func (s *Service) GetMission(ctx context.Context, missionID uuid.UUID, accountID int64) (*Mission, error) {
	mission, err := s.missionRepo.GetMission(ctx, missionID)
	if err != nil {
		return nil, apperr.Internal(ErrGetMission, "failed to get mission", err)
	}
	if err := s.validateAction(ctx, missionID, accountID, mission.IsPublic, actionView); err != nil {
		return nil, err
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

func (s *Service) GetMissionStream(ctx context.Context, missionID uuid.UUID, accountID int64) (<-chan TaskResult, func(), error) {
	mission, err := s.missionRepo.GetMission(ctx, missionID)
	if err != nil {
		return nil, nil, apperr.Internal(ErrGetMission, "failed to load mission", err)
	}
	if err := s.validateAction(ctx, missionID, accountID, mission.IsPublic, actionView); err != nil {
		return nil, nil, err
	}

	ch, cancel, err := s.progressHub.Sub(ctx, missionID)
	if err != nil {
		return nil, nil, apperr.Internal(ErrProgressHub, "failed to subscribe", err)
	}

	if mission.Map != nil && mission.DataVersion == *mission.SourceDataVersion {
		if err := s.progressHub.Pub(ctx, missionID, TaskResult{Progress: taskDone, Payload: mission.Map}); err != nil {
			cancel()
			return nil, nil, apperr.Internal(ErrProgressHub, "failed to publish", err)
		}
		return ch, cancel, nil
	}

	if err := s.enqueue(ctx, &Task{MissionID: missionID, DataVersion: mission.DataVersion, Data: &mission.Data}); err != nil {
		cancel()
		return nil, nil, apperr.Internal(ErrTaskQueue, "failed to enqueue task", err)
	}
	return ch, cancel, nil
}

func (s *Service) UpdateMissionInformation(ctx context.Context, info *Information, missionID uuid.UUID, accountID int64) error {
	if err := s.validateAction(ctx, missionID, accountID, false, actionEdit); err != nil {
		return err
	}
	if err := s.missionRepo.UpdateMissionInformation(ctx, info, missionID); err != nil {
		return apperr.Internal(ErrUpdateMission, "failed to update mission information", err)
	}
	return nil
}

func (s *Service) UpdateMissionData(ctx context.Context, data *Data, missionID uuid.UUID, accountID int64) error {
	if err := s.validateAction(ctx, missionID, accountID, false, actionEdit); err != nil {
		return err
	}
	dataVersion, err := s.missionRepo.UpdateMissionData(ctx, data, missionID)
	if err != nil {
		return apperr.Internal(ErrUpdateMission, "failed to update mission data", err)
	}
	task := &Task{
		MissionID:   missionID,
		DataVersion: dataVersion,
		Data:        data,
	}
	if err := s.enqueue(ctx, task); err != nil {
		return apperr.Internal(ErrTaskQueue, "failed to enqueue task", err)
	}
	return nil
}

func (s *Service) SetMissionMapOverrides(ctx context.Context, overrides *Overrides, missionID uuid.UUID, accountID int64) error {
	if err := s.validateAction(ctx, missionID, accountID, false, actionEdit); err != nil {
		return err
	}
	if err := s.missionRepo.SetMissionMapOverrides(ctx, overrides, missionID); err != nil {
		return apperr.Internal(ErrUpdateMission, "failed to update mission overrides", err)
	}
	return nil
}

func (s *Service) SetMissionVisibility(ctx context.Context, isPublic bool, missionID uuid.UUID, accountID int64) error {
	if err := s.validateAction(ctx, missionID, accountID, false, actionEdit); err != nil {
		return err
	}
	if err := s.missionRepo.SetMissionVisibility(ctx, isPublic, missionID); err != nil {
		return apperr.Internal(ErrUpdateMission, "failed to update mission visibility", err)
	}
	return nil
}

func (s *Service) DeleteMission(ctx context.Context, missionID uuid.UUID, accountID int64) error {
	if err := s.validateAction(ctx, missionID, accountID, false, actionEdit); err != nil {
		return err
	}
	if err := s.missionRepo.DeleteMission(ctx, missionID); err != nil {
		return apperr.Internal(ErrDeleteMission, "failed to delete mission", err)
	}
	return nil
}

func (s *Service) validateAction(ctx context.Context, missionID uuid.UUID, accountID int64, isPublic bool, action action) error {
	if isPublic && action == actionView {
		return nil
	}
	owner, err := s.missionRepo.GetMissionOwner(ctx, missionID)
	if err != nil {
		return apperr.Internal(ErrValidation, "failed to get mission owner", err)
	}

	if accountID != owner {
		return apperr.NotFound(ErrGetMission, "mission not found", err)
	}

	return nil
}
