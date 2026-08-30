package content

import (
	"astrohop/pkg/apperr"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/minio/minio-go/v7"
)

type Service struct {
	minioClient *minio.Client
}

func NewService(minioClient *minio.Client) *Service {
	return &Service{
		minioClient: minioClient,
	}
}

func (s *Service) SaveMissionMap(ctx context.Context, missionId int64, m io.Reader, size int64) error {
	objectName := fmt.Sprintf("maps/%d", missionId)
	_, err := s.minioClient.PutObject(ctx, "astrohop", objectName, m, size, minio.PutObjectOptions{})
	if err != nil {
		return apperr.New(http.StatusInternalServerError, "failed tp save mission map", err)
	}
	return nil
}

func (s *Service) GetMissionMap(ctx context.Context, missionId int64) (io.Reader, int64, error) {
	objectName := fmt.Sprintf("maps/%d", missionId)
	stat, err := s.minioClient.StatObject(ctx, "astrohop", objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, 0, apperr.New(http.StatusInternalServerError, "failed to get map stat", err)
	}
	obj, err := s.minioClient.GetObject(ctx, "astrohop", objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, apperr.New(http.StatusInternalServerError, "failed to get mission map", err)
	}

	return obj, stat.Size, nil
}
