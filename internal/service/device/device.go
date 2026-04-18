package device

import (
	"context"
	"crypto/rand"
	"technical-challenge/internal/domain/entity"
	"technical-challenge/internal/domain/gateway"
	"technical-challenge/internal/resource/database"
	"time"

	"github.com/oklog/ulid/v2"

	"go.uber.org/zap"
)

/*
TODO: tracer logging here, will be top down: from request -> service -> db -> response
It bubbles down and then up, making it easy to see where the problem is across different layers and instances
*/

var _ gateway.DeviceService = (*Service)(nil)

type Service struct {
	db    database.Service
	now   func() time.Time
	newID func() string
	log   *zap.Logger
}

func New(db database.Service, log *zap.Logger) (*Service, error) {
	if db == nil {
		return nil, entity.ErrEmptyClient
	}

	if log == nil {
		return nil, entity.ErrEmptyLogger
	}
	return &Service{
		db:    db,
		now:   func() time.Time { return time.Now().UTC() },
		newID: func() string { return ulid.MustNew(ulid.Now(), rand.Reader).String() },
		log:   log,
	}, nil
}

func (s *Service) Create(ctx context.Context, input *gateway.DeviceCreateInput) (*entity.Device, error) {
	if input == nil {
		return nil, entity.ErrEmptyDeviceCreateInput
	}
	if input.Name == "" {
		return nil, entity.ErrEmptyDeviceName
	}
	if input.Brand == "" {
		return nil, entity.ErrEmptyDeviceBrand
	}
	state := entity.StateAvailable
	if input.State != nil {
		if !input.State.Valid() {
			return nil, entity.ErrInvalidDeviceState
		}
		state = *input.State
	}

	device := entity.Device{
		ID:        s.newID(),
		Name:      input.Name,
		Brand:     input.Brand,
		State:     state,
		CreatedAt: s.now(),
		Version:   1,
	}
	err := s.db.Create(ctx, &device)
	if err != nil {
		return nil, err
	}
	return &device, nil
}
func (s *Service) List(ctx context.Context, filters *gateway.DeviceListFilter) (*entity.DevicePage, error) {
	if filters == nil {
		return nil, entity.ErrEmptyListFilter
	}

	return s.db.List(ctx, filters)
}

func (s *Service) Get(ctx context.Context, id string) (*entity.Device, error) {
	if id == "" {
		return nil, entity.ErrEmptyDeviceID
	}
	return s.db.Get(ctx, id)
}

func (s *Service) Update(ctx context.Context, id string, input *gateway.DeviceUpdateInput) (*entity.Device, error) {
	if id == "" {
		return nil, entity.ErrEmptyDeviceID
	}
	if input == nil {
		return nil, entity.ErrEmptyDeviceUpdateInput
	}
	if input.Name == "" {
		return nil, entity.ErrEmptyDeviceName
	}
	if input.Brand == "" {
		return nil, entity.ErrEmptyDeviceBrand
	}
	if input.State == nil || !input.State.Valid() {
		return nil, entity.ErrInvalidDeviceState
	}

	return s.db.Update(ctx, id, input)
}
func (s *Service) Patch(ctx context.Context, id string, input *gateway.DevicePatchInput) (*entity.Device, error) {
	if id == "" {
		return nil, entity.ErrEmptyDeviceID
	}
	if input == nil {
		return nil, entity.ErrEmptyDevicePatchInput
	}
	if input.State != nil && !input.State.Valid() {
		return nil, entity.ErrInvalidDeviceState
	}

	return s.db.Patch(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if id == "" {
		return entity.ErrEmptyDeviceID
	}
	device, err := s.db.Get(ctx, id)
	if err != nil {
		return err
	}
	if device.State == entity.StateInUse {
		return entity.ErrDeviceInUse
	}

	return s.db.Delete(ctx, id)
}
