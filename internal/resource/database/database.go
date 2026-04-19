package database

import (
	"context"
	"technical-challenge/internal/domain/entity"
	"technical-challenge/internal/domain/gateway"
)

// database contract, bound to domain interfaces

// TODO: properly define database operations
// in case of a multi type database, where there are no only devices, but more entries for the same ops, would go with generics
// for now, will be only for devices tho
type Service interface {
	Create(ctx context.Context, device *entity.Device) error
	List(ctx context.Context, filter *gateway.DeviceListFilter) (*entity.DevicePage, error)
	Get(ctx context.Context, id string) (*entity.Device, error)
	Update(ctx context.Context, id string, input *gateway.DeviceUpdateInput) (*entity.Device, error)
	Patch(ctx context.Context, id string, input *gateway.DevicePatchInput) (*entity.Device, error)
	Delete(ctx context.Context, id string) error
	Close(ctx context.Context) error
}
