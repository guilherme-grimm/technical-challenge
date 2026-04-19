package device

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"

	"technical-challenge/internal/domain/entity"
	"technical-challenge/internal/domain/gateway"
	"technical-challenge/internal/logger"
	"technical-challenge/internal/resource/database"
)

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
	log := logger.FromContext(ctx, s.log).With(zap.String("op", "service.Create"))
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
	if err := s.db.Create(ctx, &device); err != nil {
		log.Error("db create failed", zap.Error(err))
		return nil, err
	}
	return &device, nil
}

func (s *Service) List(ctx context.Context, filters *gateway.DeviceListFilter) (*entity.DevicePage, error) {
	log := logger.FromContext(ctx, s.log).With(zap.String("op", "service.List"))
	if filters == nil {
		return nil, entity.ErrEmptyListFilter
	}
	page, err := s.db.List(ctx, filters)
	if err != nil {
		log.Error("db list failed", zap.Error(err))
		return nil, err
	}
	return page, nil
}

func (s *Service) Get(ctx context.Context, id string) (*entity.Device, error) {
	log := logger.FromContext(ctx, s.log).With(zap.String("op", "service.Get"), zap.String("id", id))
	if id == "" {
		return nil, entity.ErrEmptyDeviceID
	}
	d, err := s.db.Get(ctx, id)
	if err != nil {
		log.Error("db get failed", zap.Error(err))
		return nil, err
	}
	return d, nil
}

func (s *Service) Update(ctx context.Context, id string, input *gateway.DeviceUpdateInput) (*entity.Device, error) {
	log := logger.FromContext(ctx, s.log).With(zap.String("op", "service.Update"), zap.String("id", id))
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

	current, err := s.db.Get(ctx, id)
	if err != nil {
		log.Error("db get (pre-update) failed", zap.Error(err))
		return nil, err
	}
	if current.State == entity.StateInUse && (current.Name != input.Name || current.Brand != input.Brand) {
		return nil, entity.ErrDeviceInUseImmutable
	}

	updated, err := s.db.Update(ctx, id, input)
	if err != nil {
		log.Error("db update failed", zap.Error(err))
		return nil, err
	}
	return updated, nil
}

func (s *Service) Patch(ctx context.Context, id string, input *gateway.DevicePatchInput) (*entity.Device, error) {
	log := logger.FromContext(ctx, s.log).With(zap.String("op", "service.Patch"), zap.String("id", id))
	if id == "" {
		return nil, entity.ErrEmptyDeviceID
	}
	if input == nil {
		return nil, entity.ErrEmptyDevicePatchInput
	}
	if input.State != nil && !input.State.Valid() {
		return nil, entity.ErrInvalidDeviceState
	}

	current, err := s.db.Get(ctx, id)
	if err != nil {
		log.Error("db get (pre-patch) failed", zap.Error(err))
		return nil, err
	}
	if current.State == entity.StateInUse {
		if input.Name != nil && *input.Name != current.Name {
			return nil, entity.ErrDeviceInUseImmutable
		}
		if input.Brand != nil && *input.Brand != current.Brand {
			return nil, entity.ErrDeviceInUseImmutable
		}
	}

	updated, err := s.db.Patch(ctx, id, input)
	if err != nil {
		log.Error("db patch failed", zap.Error(err))
		return nil, err
	}
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	log := logger.FromContext(ctx, s.log).With(zap.String("op", "service.Delete"), zap.String("id", id))
	if id == "" {
		return entity.ErrEmptyDeviceID
	}
	current, err := s.db.Get(ctx, id)
	if err != nil {
		log.Error("db get (pre-delete) failed", zap.Error(err))
		return err
	}
	if current.State == entity.StateInUse {
		return entity.ErrDeviceInUse
	}
	if err := s.db.Delete(ctx, id); err != nil {
		log.Error("db delete failed", zap.Error(err))
		return err
	}
	return nil
}
