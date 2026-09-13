package services

import (
	"context"
	"errors"
	"fmt"

	"media-sequencer-backend/internal/models"
	"media-sequencer-backend/internal/repository"
)

type MediaService interface {
	GetAll(ctx context.Context) ([]models.Media, error)
	GetByID(ctx context.Context, id int) (*models.Media, error)
	Create(ctx context.Context, req models.CreateMediaRequest) (*models.Media, error)
	Delete(ctx context.Context, id int) error
}

type mediaService struct {
	mediaRepo repository.MediaRepository
}

func NewMediaService(mediaRepo repository.MediaRepository) MediaService {
	return &mediaService{
		mediaRepo: mediaRepo,
	}
}

func (s *mediaService) GetAll(ctx context.Context) ([]models.Media, error) {
	return s.mediaRepo.GetAll(ctx)
}

func (s *mediaService) GetByID(ctx context.Context, id int) (*models.Media, error) {
	media, err := s.mediaRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch media: %w", err)
	}
	if media == nil {
		return nil, errors.New("media item not found")
	}
	return media, nil
}

func (s *mediaService) Create(ctx context.Context, req models.CreateMediaRequest) (*models.Media, error) {
	if req.Duration <= 0 {
		return nil, errors.New("duration must be greater than zero")
	}
	if req.Type != models.MediaTypeImage && req.Type != models.MediaTypeVideo && req.Type != models.MediaTypeBlank {
		return nil, errors.New("invalid media type: must be 'image', 'video', or 'blank'")
	}
	return s.mediaRepo.Create(ctx, req)
}

func (s *mediaService) Delete(ctx context.Context, id int) error {
	return s.mediaRepo.Delete(ctx, id)
}
