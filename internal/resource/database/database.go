package database

import (
	"context"
	"technical-challenge/internal/domain/entity"
	"technical-challenge/internal/domain/gateway"
)

type Service interface {
	Create(ctx context.Context, device *entity.Device) error
	List(ctx context.Context, filter *gateway.DeviceListFilter) (*entity.DevicePage, error)
	Get(ctx context.Context, id string) (*entity.Device, error)
	Update(ctx context.Context, id string, input *gateway.DeviceUpdateInput) (*entity.Device, error)
	Patch(ctx context.Context, id string, input *gateway.DevicePatchInput) (*entity.Device, error)
	Delete(ctx context.Context, id string) error
	Close(ctx context.Context) error
	Ping(ctx context.Context) error
}
