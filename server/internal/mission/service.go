package mission

import (
	"astrohop/internal/astronomy/atime"
	"astrohop/internal/astronomy/coordinates"
	"astrohop/internal/astronomy/sol"
	"astrohop/internal/search"
	"astrohop/pkg/algo"
	"astrohop/pkg/apperr"
	"context"

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
	Queue(ctx context.Context) (chan *Task, error)
	Enqueue(ctx context.Context, task *Task) error
	SetTaskProgressChan(ctx context.Context, missionID uuid.UUID, progress chan TaskResult) error
	GetTaskProgressChan(ctx context.Context, missionID uuid.UUID) (chan TaskResult, error)
}

type Service struct {
	missionRepo Repo
	taskQueue   TaskQueue
	searchSvc   *search.Service
}

func NewService(
	missionRepo Repo,
	taskQueue TaskQueue,
	searchSvc *search.Service,
) *Service {
	return &Service{
		searchSvc:   searchSvc,
		missionRepo: missionRepo,
		taskQueue:   taskQueue,
	}
}

func (s *Service) Start(ctx context.Context) error {
	q, err := s.taskQueue.Queue(ctx)
	if err != nil {
		return err
	}
	for range 5 {
		go s.missionWorker(ctx, q)
	}
	return nil
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

func (s *Service) GetMissionStream(ctx context.Context, id uuid.UUID) (<-chan TaskResult, error) {
	ch, err := s.taskQueue.GetTaskProgressChan(ctx, id)
	if err != nil {
		return nil, apperr.Internal(ErrTaskQueue, "failed to get task progress channel", err)
	}

	if ch != nil {
		return ch, nil
	}

	mission, err := s.missionRepo.GetMission(ctx, id)
	if err != nil {
		return nil, apperr.Internal(ErrGetMission, "failed to get mission task status stream", err)
	}

	if mission.MapData == nil {
		task := &Task{
			MissionID: id,
			Data:      mission.Data,
		}
		if err := s.taskQueue.Enqueue(ctx, task); err != nil {
			if err = s.taskQueue.SetTaskProgressChan(ctx, id, nil); err != nil {
				return nil, apperr.Internal(ErrTaskQueue, "failed to set task progress chan", err)
			}
			return nil, apperr.Internal(ErrTaskQueue, "failed to enqueue task", err)
		}
		ch, err := s.taskQueue.GetTaskProgressChan(ctx, id)
		if err != nil {
			return nil, apperr.Internal(ErrTaskQueue, "failed to get task progress channel", err)
		}
		return ch, nil
	}

	ch = make(chan TaskResult)
	go func() {
		ch <- TaskResult{
			Progress: taskDone,
			Payload:  mission.MapData,
		}
		close(ch)
	}()
	return ch, nil
}

func (s *Service) missionWorker(ctx context.Context, q chan *Task) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-q:
			func() {
				progressChan, err := s.taskQueue.GetTaskProgressChan(ctx, task.MissionID)
				if err != nil {
					//TOO: error handling
					return
				}
				defer func() {
					err := s.taskQueue.SetTaskProgressChan(ctx, task.MissionID, nil)
					if err != nil {
						//TOO: error handling
						return
					}
					close(progressChan)
				}()

				progressChan <- TaskResult{Progress: taskBuildingRoute}

				lst := atime.GetLocalSidereal(task.Data.Location, task.Data.Time)
				objectives := task.Data.Objectives
				objectOids := make([]int64, len(objectives))
				for i, objective := range objectives {
					objectOids[i] = objective.OID
				}

				objectStellarData, err := s.searchSvc.GetObjectsStellarData(ctx, objectOids)

				positions := make(map[int64]coordinates.Horizontal, len(objectives))
				for i, d := range objectStellarData {
					positions[objectOids[i]] = d.EqCoords.ToHorizontal(lst, task.Data.Location.Lat)
				}

				tour, err := s.buildTour(objectOids, objectStellarData)
				if err != nil {
					progressChan <- TaskResult{Progress: taskFailed, Error: err}
					return
				}

				moonPosition, _ := sol.CalculateMoonPosition(task.Data.Time)

				mapData := &MapData{
					MoonPosition: moonPosition.ToEquatorial().ToHorizontal(lst, task.Data.Location.Lat),
					Positions:    positions,
					Tour:         tour,
				}

				if err := s.missionRepo.UpdateMissionMapData(ctx, task.MissionID, mapData); err != nil {
					progressChan <- TaskResult{Progress: taskFailed, Error: err}
					return
				}

				progressChan <- TaskResult{Progress: taskDone, Payload: mapData}
			}()
		}
	}
}

func (s *Service) buildTour(objectOids []int64, stellarData []search.ObjectStellarData) ([]int64, error) {
	distanceMtrx := make([][]float64, len(stellarData))
	for i := range stellarData {
		distanceMtrx[i] = make([]float64, len(stellarData))
	}
	for i := 0; i < len(stellarData)-1; i++ {
		for j := i + 1; j < len(stellarData); j++ {
			dist := coordinates.DistanceEq(stellarData[i].EqCoords, stellarData[j].EqCoords)
			distanceMtrx[i][j] = dist
			distanceMtrx[j][i] = dist
		}
	}
	order := algo.UseFarthestInsertion(distanceMtrx)
	tour := make([]int64, len(objectOids))
	for i, o := range order {
		tour[i] = objectOids[o]
	}
	return tour, nil
}
