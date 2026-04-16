package device

import (
	"context"
	"technical-challenge/internal/domain/entity"
	"technical-challenge/internal/domain/gateway"
)

// NOTE: stubs for now, laying out the place

var _ gateway.DeviceService = (*Service)(nil)

type Service struct {
	// db interface
}

func (s *Service) Create(ctx context.Context) (*entity.Device, error) {

	return nil, nil
}
func (s *Service) List(ctx context.Context) ([]entity.PaginatedResponse, error) {

	return nil, nil
}
func (s *Service) Get(ctx context.Context) (*entity.Device, error) {

	return nil, nil
}
func (s *Service) Update(ctx context.Context) (*entity.Device, error) {

	return nil, nil
}
func (s *Service) Patch(ctx context.Context) (*entity.Device, error) {

	return nil, nil
}
func (s *Service) Delete(ctx context.Context) (*entity.Device, error) {

	return nil, nil
}
