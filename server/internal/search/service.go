package search

import (
	"astrohop/internal/astronomy/coordinates"
	"astrohop/pkg/apperr"
	"context"
	"net/http"
	"strings"
)

type Repo interface {
	SearchObjectsByName(ctx context.Context, name string) ([]Object, error)
	GetObjectNavData(ctx context.Context, oid int64) (*NavData, error)
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
		return nil, apperr.New(http.StatusInternalServerError, "failed to find objects by name", err)
	}
	return objects, nil
}

func (s *Service) GetObjectStellarData(ctx context.Context, oid int64) (*ObjectStellarData, error) {
	data, err := s.repo.GetObjectNavData(ctx, oid)
	if err != nil {
		return nil, apperr.New(http.StatusInternalServerError, "failed to find object stellar data", err)
	}
	return &ObjectStellarData{
		EqCoords: coordinates.Equatorial{
			RA:  data.RA,
			Dec: data.Dec,
		},
		ApparentMagnitude: data.ApparentMagnitude,
	}, nil
}

func normalizeObjectName(name string) string {
	return strings.ToLower(strings.Join(strings.Split(name, " "), ""))
}
