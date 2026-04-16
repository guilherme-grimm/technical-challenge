package mongo

import (
	"context"
	"technical-challenge/internal/domain/entity"
	"technical-challenge/internal/resource/database"
)

// mongo things will be here

var _ database.Service = (*Service)(nil)

type Service struct {
}

func (s *Service) Create(ctx context.Context, device entity.Device) error {
	return nil
}
func (s *Service) List(ctx context.Context) ([]entity.Device, error) {
	return nil, nil
}
func (s *Service) Get(ctx context.Context) (*entity.Device, error) {
	return nil, nil
}
func (s *Service) Update(ctx context.Context) error {
	return nil
}
func (s *Service) Patch(ctx context.Context) error {
	return nil
}
func (s *Service) Delete(ctx context.Context) error {
	return nil
}
