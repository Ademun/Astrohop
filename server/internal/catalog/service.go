package catalog

import (
	"astrohop/pkg/apperr"
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	searchLimit     = 10
	defaultPageSize = 50
	maxPageSize     = 200
)

type Repo interface {
	Collections(ctx context.Context) ([]Collection, error)
	CollectionObjects(ctx context.Context, id CollectionID, page PageRequest) (Page[CollectionMember], error)
	Object(ctx context.Context, id ObjectID) (Object, error)
	Positions(ctx context.Context, ids []ObjectID) ([]ObjectPosition, error)
	Search(ctx context.Context, query string, limit int) ([]SearchHit, error)
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

func (s *Service) CollectionObjects(ctx context.Context, id CollectionID, page PageRequest) (Page[CollectionMember], error) {
	if page.Limit == 0 {
		page.Limit = defaultPageSize
	}
	if page.Limit < 0 || page.Limit > maxPageSize || page.Offset < 0 {
		return Page[CollectionMember]{}, apperr.Validation(
			ErrInvalidPage,
			fmt.Sprintf("limit must be between 1 and %d, offset must not be negative", maxPageSize),
			nil,
		)
	}

	res, err := s.repo.CollectionObjects(ctx, id, page)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Page[CollectionMember]{}, apperr.NotFound(ErrCollectionNotFound, "collection not found", err)
		}
		return Page[CollectionMember]{}, apperr.Internal(ErrCollectionObjects, "failed to get collection objects", err)
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

func (s *Service) Search(ctx context.Context, query string) ([]SearchHit, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, apperr.Validation(ErrSearchEmptyQuery, "search query must not be empty", nil)
	}
	hits, err := s.repo.Search(ctx, query, searchLimit)
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
