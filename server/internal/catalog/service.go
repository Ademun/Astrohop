package catalog

import (
	"astrohop/pkg/apperr"
	"context"
	"errors"
)

const (
	defaultSearchSize = 10
	maxSearchSize     = 200
)

type Repo interface {
	Collections(ctx context.Context) ([]Collection, error)
	Object(ctx context.Context, id ObjectID) (Object, error)
	Positions(ctx context.Context, ids []ObjectID) ([]ObjectPosition, error)
	Search(ctx context.Context, request SearchRequest) ([]SearchHit, error)
	ConstellationByStar(ctx context.Context, id ObjectID) (*Constellation, error)
}

type Service struct {
	repo Repo
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) Collections(ctx context.Context) ([]Collection, error) {
	res, err := s.repo.Collections(ctx)
	if err != nil {
		return nil, apperr.Internal(ErrListCollections, "failed to list collections", err)
	}
	return res, nil
}

func (s *Service) Object(ctx context.Context, id ObjectID) (Object, error) {
	obj, err := s.repo.Object(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Object{}, apperr.NotFound(ErrObjectNotFound, "object not found", err)
		}
		return Object{}, apperr.Internal(ErrGetObject, "failed to get object", err)
	}
	return obj, nil
}

func (s *Service) Positions(ctx context.Context, ids []ObjectID) ([]ObjectPosition, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	res, err := s.repo.Positions(ctx, ids)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, apperr.NotFound(ErrObjectNotFound, "one or more objects not found", err)
		}
		return nil, apperr.Internal(ErrGetPositions, "failed to get object positions", err)
	}
	return res, nil
}

func (s *Service) Search(ctx context.Context, request SearchRequest) ([]SearchHit, error) {
	if request.Limit <= 0 {
		request.Limit = defaultSearchSize
	}
	if request.Limit > maxSearchSize {
		request.Limit = maxSearchSize
	}
	hits, err := s.repo.Search(ctx, request)
	if err != nil {
		return nil, apperr.Internal(ErrSearch, "failed to search objects", err)
	}
	return hits, nil
}

func (s *Service) ConstellationOf(ctx context.Context, id ObjectID) (*Constellation, error) {
	res, err := s.repo.ConstellationByStar(ctx, id)
	if err != nil {
		return nil, apperr.Internal(ErrConstellationsOf, "failed to get constellation of object", err)
	}
	return res, nil
}
