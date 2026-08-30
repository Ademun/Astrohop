package planner

import (
	"astrohop/internal/astronomy/atime"
	"astrohop/internal/astronomy/coordinates"
	"astrohop/internal/content"
	"astrohop/internal/plotter"
	"astrohop/internal/search"
	"astrohop/pkg/algo"
	"astrohop/pkg/apperr"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
)

type MissionRepo interface {
	CreateNewMission(ctx context.Context, mission *MissionInfo, accessToken string) (int64, error)
	GetMission(ctx context.Context, missionId int64) (*Mission, error)
}

type Service struct {
	missionRepo MissionRepo
	searchSvc   *search.Service
	contentSvc  *content.Service
	plotter     *plotter.SvgPlotter
	minioClient *minio.Client
	tasks       chan missionTask
}

func NewService(
	missionRepo MissionRepo,
	searchSvc *search.Service,
	contentSvc *content.Service,
	minioClient *minio.Client,
) *Service {
	return &Service{
		searchSvc:   searchSvc,
		contentSvc:  contentSvc,
		missionRepo: missionRepo,
		plotter:     &plotter.SvgPlotter{},
		minioClient: minioClient,
		tasks:       make(chan missionTask),
	}
}

func (s *Service) Start(ctx context.Context) {
	for range 5 {
		go s.missionTaskWorker(ctx)
	}
}

func (s *Service) PreviewMission(ctx context.Context, m MissionInfo) (*MissionValidationInfo, string, error) {
	validation, err := s.validateMission(ctx, &m)
	if err != nil {
		return nil, "", apperr.New(http.StatusInternalServerError, "failed to validate mission", err)
	}
	token, err := generateAccessToken()
	if err != nil {
		return nil, "", apperr.New(http.StatusInternalServerError, "failed to generate access token", err)
	}
	missionID, err := s.missionRepo.CreateNewMission(ctx, &m, token)
	if err != nil {
		return nil, "", apperr.New(http.StatusInternalServerError, "failed to create mission task", err)
	}
	validation.MissionID = missionID
	return validation, token, nil
}

func (s *Service) ProcessMission(missionId int64) (chan taskResult, error) {
	task := missionTask{
		MissionID: missionId,
		Result:    make(chan taskResult),
	}

	select {
	case <-time.After(10 * time.Second):
		return nil, apperr.New(http.StatusRequestTimeout, "task enqueue timeout", nil)
	case s.tasks <- task:
		return task.Result, nil
	}
}

func (s *Service) missionTaskWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-s.tasks:
			func() {
				defer close(task.Result)
				task.Result <- taskResult{
					Progress: taskStarted,
				}
				mission, err := s.missionRepo.GetMission(ctx, task.MissionID)
				if err != nil {
					task.Result <- taskResult{
						Progress: taskFailed,
						Error:    err,
					}
					return
				}
				task.Result <- taskResult{
					Progress: taskFetchingData,
				}
				objectives := mission.Info.Objectives

				coords := make(map[int]coordinates.Equatorial)
				for i, obj := range objectives {
					sd, err := s.searchSvc.GetObjectStellarData(ctx, obj.OID)
					if err != nil {
						task.Result <- taskResult{
							Progress: taskFailed,
							Error:    err,
						}
						return
					}
					coords[i] = sd.EqCoords
				}

				distanceMtrx := make([][]float64, len(objectives))
				for i := range objectives {
					distanceMtrx[i] = make([]float64, len(objectives))
					for j := range objectives {
						if i == j {
							continue
						}
						dist := coordinates.DistanceEq(coords[i], coords[j])
						distanceMtrx[i][j] = dist
					}
				}
				task.Result <- taskResult{
					Progress: taskBuildingRoute,
				}
				tour := algo.UseFarthestInsertion(distanceMtrx)

				lst := atime.GetLocalSidereal(mission.Info.Location, mission.Info.Time)
				chartObjects := make([]plotter.Object, len(objectives))
				for i, obj := range objectives {
					chartObjects[i] = plotter.Object{
						Label:  obj.Name,
						Coords: coords[i].ToHorizontal(lst, mission.Info.Location.Lat),
					}
				}
				plotData := &plotter.ChartData{
					Title:    "ASTROHOP",
					Location: mission.Info.Location,
					Time:     mission.Info.Time,
					Conditions: plotter.ChartConditions{
						Bortle:            6,
						Equipment:         "Binoculars 10x50",
						LimitingMagnitude: 6.0,
					},
					Objects:     chartObjects,
					Tour:        tour,
					AngularTips: plotter.DefaultAngularTips(),
					ArmNote:     "* Extend your arm as far as possible",
					Legend:      plotter.DefaultLegend(),
				}
				var buf bytes.Buffer
				gz := gzip.NewWriter(&buf)
				task.Result <- taskResult{
					Progress: taskGeneratingMap,
				}
				s.plotter.PlotAzimuth(gz, plotter.OptionsDefault, plotData)
				if err := gz.Close(); err != nil {
					task.Result <- taskResult{
						Progress: taskFailed,
						Error:    err,
					}
					return
				}
				if err := s.contentSvc.SaveMissionMap(ctx, task.MissionID, &buf, int64(buf.Len())); err != nil {
					task.Result <- taskResult{
						Progress: taskFailed,
						Error:    err,
					}
				}

				task.Result <- taskResult{
					Progress: taskDone,
					Payload:  buf.Bytes(),
				}
			}()
		}
	}
}

func (s *Service) validateMission(ctx context.Context, m *MissionInfo) (*MissionValidationInfo, error) {
	result := &MissionValidationInfo{
		Objectives: make([]MissionValidatedObjective, len(m.Objectives)),
	}
	lst := atime.GetLocalSidereal(m.Location, m.Time)
	if len(m.Objectives) == 0 {
		return nil, apperr.New(http.StatusUnprocessableEntity, "objective list cannot be empty", errors.New("objective list cannot be empty"))
	}
	for i, o := range m.Objectives {
		stellarData, err := s.searchSvc.GetObjectStellarData(ctx, o.OID)
		if err != nil {
			return nil, err
		}
		validateVisibility(result, i, o, stellarData.ApparentMagnitude, m.Conditions.LimitingMagnitude)
		validatePosition(result, i, stellarData.EqCoords, lst, m.Location.Lat)
	}
	return result, nil
}

func validateVisibility(vi *MissionValidationInfo, i int, o MissionObjective, om, lm float32) {
	magDiff := om - lm
	if magDiff < 1.0 {
		vi.Objectives[i] = MissionValidatedObjective{
			Name:              o.Name,
			ApparentMagnitude: om,
			Status:            ObjectiveStatusVisible,
		}
	} else if magDiff >= 1.0 && magDiff < 2.0 {
		vi.Objectives[i] = MissionValidatedObjective{
			Name:              o.Name,
			ApparentMagnitude: om,
			Status:            ObjectiveStatusDim,
		}
	} else {
		vi.Objectives[i] = MissionValidatedObjective{
			Name:              o.Name,
			ApparentMagnitude: om,
			Status:            ObjectiveStatusInvisible,
		}
	}
}

func validatePosition(vi *MissionValidationInfo, i int, eqCoords coordinates.Equatorial, lst float64, lat float64) {
	hCoords := eqCoords.ToHorizontal(lst, lat)
	if hCoords.Alt < 0 {
		vi.Objectives[i].Status = ObjectiveStatusInvisible
	}
}

func generateAccessToken() (string, error) {
	aes := make([]byte, 32)
	if _, err := rand.Read(aes); err != nil {
		return "", err
	}
	aesBase := base64.URLEncoding.EncodeToString(aes)
	return aesBase, nil
}
