package search

import (
	"astrohop/internal/astronomy/coordinates"
	"astrohop/pkg/apperr"
	"context"
	"strings"
)

type Repo interface {
	SearchObjectsByName(ctx context.Context, name string) ([]Object, error)
	GetObjectsNavData(ctx context.Context, oid []int64) ([]NavData, error)
}

type Service struct {
	repo Repo
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) SearchObjectsByName(ctx context.Context, name string) ([]Object, error) {
	norm := normalizeObjectName(name)
	objects, err := s.repo.SearchObjectsByName(ctx, norm)
	if err != nil {
		return nil, apperr.Internal(ErrSearch, "failed to search objects by name", err)
	}
	return objects, nil
}

func (s *Service) GetObjectsStellarData(ctx context.Context, oid []int64) ([]ObjectStellarData, error) {
	data, err := s.repo.GetObjectsNavData(ctx, oid)
	if err != nil {
		return nil, apperr.Internal(ErrGetObject, "failed to get objects nav data", err)
	}
	stellarData := make([]ObjectStellarData, len(data))
	for i, obj := range data {
		stellarData[i] = ObjectStellarData{
			EqCoords: coordinates.Equatorial{
				RA:  obj.RA,
				Dec: obj.Dec,
			},
			ApparentMagnitude: obj.ApparentMagnitude,
		}
	}
	return stellarData, nil
}

func normalizeObjectName(name string) string {
	return strings.ToLower(strings.Join(strings.Split(name, " "), ""))
}
