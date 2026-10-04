package account

import (
	"astrohop/pkg/apperr"
	"astrohop/pkg/utils"
	"context"
	"encoding/base64"
)

type Repo interface {
	CreateAccount(ctx context.Context, key string) error
}

type Service struct {
	repo Repo
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateAccount(ctx context.Context) (string, error) {
	key, err := utils.RandomKey(256)
	if err != nil {
		return "", apperr.Internal(ErrGenKey, "failed to generate account key", err)
	}
	urlKey := base64.URLEncoding.EncodeToString(key)

	if err := s.repo.CreateAccount(ctx, urlKey); err != nil {
		return "", apperr.Internal(ErrCreateAcc, "failed to create account", err)
	}

	return urlKey, nil
}
