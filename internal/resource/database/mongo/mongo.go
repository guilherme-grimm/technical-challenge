package mongo

import (
	"context"
	"technical-challenge/internal/domain/entity"
	"technical-challenge/internal/domain/gateway"
	"technical-challenge/internal/resource/database"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

var _ database.Service = (*Service)(nil)

type Service struct {
	client     *mongo.Client
	collection *mongo.Collection
	log        *zap.Logger
}

func (s *Service) Create(ctx context.Context, device *entity.Device) error {
	return nil
}
func (s *Service) List(ctx context.Context, filter *gateway.DeviceListFilter) (*entity.DevicePage, error) {
	return nil, nil
}
func (s *Service) Get(ctx context.Context, id string) (*entity.Device, error) {
	return nil, nil
}
func (s *Service) Update(ctx context.Context, id string, input *gateway.DeviceUpdateInput) (*entity.Device, error) {
	return nil, nil
}
func (s *Service) Patch(ctx context.Context, id string, input *gateway.DevicePatchInput) (*entity.Device, error) {
	return nil, nil
}
func (s *Service) Delete(ctx context.Context, id string) error {
	return nil
}
