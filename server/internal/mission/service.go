package mission

import (
	"astrohop/internal/astronomy/coordinates"
	"astrohop/internal/search"
	"astrohop/pkg/algo"
	"astrohop/pkg/apperr"
	"context"
	"net/http"

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

type Service struct {
	missionRepo Repo
	searchSvc   *search.Service
	pool        *pool
}

func NewService(
	missionRepo Repo,
	searchSvc *search.Service,
) *Service {
	return &Service{
		searchSvc:   searchSvc,
		missionRepo: missionRepo,
		pool:        newPool(),
	}
}

func (s *Service) Start(ctx context.Context) {
	for range 5 {
		go s.missionWorker(ctx, s.pool.Queue)
	}
}

func (s *Service) CreateMission(ctx context.Context, data *Data, accountID int64) (uuid.UUID, error) {
	missionID, err := s.missionRepo.CreateMission(ctx, data, accountID)
	if err != nil {
		return uuid.Nil, apperr.New(http.StatusInternalServerError, "mission.service: failed to create new mission", err)
	}
	task := missionTask{
		MissionID: missionID,
		Data:      data,
	}
	if err := s.pool.enqueueTask(task); err != nil {
		return uuid.Nil, apperr.New(http.StatusInternalServerError, "mission.service: failed to enqueue task", err)
	}
	return missionID, nil
}

func (s *Service) UpdateMissionData(ctx context.Context, missionID uuid.UUID, data *Data) error {
	if err := s.missionRepo.UpdateMissionData(ctx, missionID, data); err != nil {
		return apperr.New(http.StatusInternalServerError, "mission.service: failed to update data", err)
	}
	task := missionTask{
		MissionID: missionID,
		Data:      data,
	}
	if err := s.pool.enqueueTask(task); err != nil {
		return apperr.New(http.StatusInternalServerError, "mission.service: failed to enqueue task", err)
	}
	return nil
}

func (s *Service) UpdateMissionVisibility(ctx context.Context, missionID uuid.UUID, isPublic bool) error {
	if err := s.missionRepo.UpdateMissionVisibility(ctx, missionID, isPublic); err != nil {
		return apperr.New(http.StatusInternalServerError, "mission.service: failed to update visibility", err)
	}
	return nil
}

func (s *Service) DeleteMission(ctx context.Context, missionID uuid.UUID) error {
	if err := s.missionRepo.DeleteMission(ctx, missionID); err != nil {
		return apperr.New(http.StatusInternalServerError, "mission.service: failed to delete mission", err)
	}
	return nil
}

func (s *Service) GetMission(ctx context.Context, id uuid.UUID) (*Mission, error) {
	mission, err := s.missionRepo.GetMission(ctx, id)
	if err != nil {
		return nil, apperr.New(http.StatusInternalServerError, "mission.service: failed to get mission", err)
	}
	return mission, nil
}

func (s *Service) GetAccountMissions(ctx context.Context, accountID int64) ([]Mission, error) {
	missions, err := s.missionRepo.GetAccountMissions(ctx, accountID)
	if err != nil {
		return nil, apperr.New(http.StatusInternalServerError, "mission.service: failed to get missions", err)
	}
	return missions, nil
}

func (s *Service) GetMissionStream(ctx context.Context, id uuid.UUID) (<-chan taskResult, error) {
	if ch := s.pool.getTaskProgressChan(id); ch != nil {
		return ch, nil
	}

	mission, err := s.missionRepo.GetMission(ctx, id)
	if err != nil {
		return nil, apperr.New(http.StatusInternalServerError, "mission.service: failed to get mission", err)
	}

	if mission.MapData == nil {
		task := missionTask{
			MissionID: id,
			Data:      mission.Data,
		}
		if err := s.pool.enqueueTask(task); err != nil {
			s.pool.setTaskProgressChan(id, nil)
			return nil, apperr.New(http.StatusInternalServerError, "mission.service: failed to enqueue task", err)
		}
		return s.pool.getTaskProgressChan(id), nil
	}

	ch := make(chan taskResult)
	go func() {
		ch <- taskResult{
			Progress: taskDone,
			Payload:  mission.MapData,
		}
		close(ch)
	}()
	return ch, nil
}

func (s *Service) missionWorker(ctx context.Context, q chan missionTask) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-q:
			func() {
				progressChan := s.pool.getTaskProgressChan(task.MissionID)
				defer func() {
					s.pool.setTaskProgressChan(task.MissionID, nil)
					close(progressChan)
				}()

				progressChan <- taskResult{Progress: taskBuildingRoute}

				objectives := task.Data.Objectives
				tour, err := s.buildTour(ctx, objectives)
				if err != nil {
					progressChan <- taskResult{Progress: taskFailed, Error: err}
					return
				}

				mapData := &MapData{Tour: tour}
				if err := s.missionRepo.UpdateMissionMapData(ctx, task.MissionID, mapData); err != nil {
					progressChan <- taskResult{Progress: taskFailed, Error: err}
					return
				}

				progressChan <- taskResult{Progress: taskDone, Payload: mapData}
			}()
		}
	}
}

func (s *Service) buildTour(ctx context.Context, objectives []Objective) (*algo.Tour, error) {
	objectOids := make([]int64, len(objectives))
	for i, objective := range objectives {
		objectOids[i] = objective.OID
	}
	objectStellarData, err := s.searchSvc.GetObjectsStellarData(ctx, objectOids)
	if err != nil {
		return nil, err
	}
	distanceMtrx := make([][]float64, len(objectives))
	for i := range objectives {
		distanceMtrx[i] = make([]float64, len(objectives))
	}
	for i := 0; i < len(objectStellarData)-1; i++ {
		for j := i + 1; j < len(objectStellarData); j++ {
			dist := coordinates.DistanceEq(objectStellarData[i].EqCoords, objectStellarData[j].EqCoords)
			distanceMtrx[i][j] = dist
			distanceMtrx[j][i] = dist
		}
	}
	tour := algo.UseFarthestInsertion(distanceMtrx)
	return tour, nil
}
