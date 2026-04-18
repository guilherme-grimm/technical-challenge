package gateway

import (
	"context"
	"technical-challenge/internal/domain/entity"
)

type DeviceCreateInput struct {
	Name  string
	Brand string
	State *entity.State
}

// full update, needs everything
type DeviceUpdateInput struct {
	Name    string
	Brand   string
	State   *entity.State
	Version int64
}

// partial, with optional fields as pointers (which pains me)
type DevicePatchInput struct {
	Name    *string
	Brand   *string
	State   *entity.State
	Version int64
}

type DeviceListFilter struct {
	Brand  *string
	State  *entity.State
	Cursor string
	Limit  int
}

type DeviceService interface {
	Create(ctx context.Context, input *DeviceCreateInput) (*entity.Device, error)
	List(ctx context.Context, filters *DeviceListFilter) (*entity.DevicePage, error)
	Get(ctx context.Context, id string) (*entity.Device, error)
	Update(ctx context.Context, id string, input *DeviceUpdateInput) (*entity.Device, error)
	Patch(ctx context.Context, id string, input *DevicePatchInput) (*entity.Device, error)
	Delete(ctx context.Context, id string) error
}
