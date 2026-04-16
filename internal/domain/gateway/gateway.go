package gateway

import (
	"context"
	"technical-challenge/internal/domain/entity"
)

// act as the contracts gateway, so the whole services that are exposed to the handler/api are defined here

// TODO: Add the methods
type DeviceService interface {
	Create(ctx context.Context) (*entity.Device, error)
	List(ctx context.Context) ([]entity.PaginatedResponse, error)
	Get(ctx context.Context) (*entity.Device, error)
	Update(ctx context.Context) (*entity.Device, error)
	Patch(ctx context.Context) (*entity.Device, error)
	Delete(ctx context.Context) (*entity.Device, error)
}
