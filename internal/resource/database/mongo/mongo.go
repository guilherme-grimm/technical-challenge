package mongo

import (
	"context"
	"errors"
	"technical-challenge/internal/domain/entity"
	"technical-challenge/internal/domain/gateway"
	"technical-challenge/internal/logger"
	"technical-challenge/internal/resource/database"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

// NOTE: No need to validate input, validate in the upper layer

var _ database.Service = (*Service)(nil)

type Service struct {
	collection *mongo.Collection
	log        *zap.Logger
}

func New(client *mongo.Client, database, collection string, log *zap.Logger) (*Service, error) {
	if client == nil {
		return nil, entity.ErrEmptyClient
	}

	if log == nil {
		return nil, entity.ErrEmptyLogger
	}
	return &Service{
		collection: client.Database(database).Collection(collection),
		log:        log,
	}, nil
}

func (s *Service) Create(ctx context.Context, device *entity.Device) error {
	log := logger.FromContext(ctx, s.log).With(
		zap.String("op", "mongo.Create"),
		zap.String("id", device.ID),
	)
	_, err := s.collection.InsertOne(ctx, device)
	if err != nil {
		log.Error("failed to insert device", zap.Error(err))
		return err
	}
	return nil
}

func (s *Service) Get(ctx context.Context, id string) (*entity.Device, error) {
	log := logger.FromContext(ctx, s.log).With(
		zap.String("op", "mongo.Get"),
		zap.String("id", id),
	)
	var device entity.Device

	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&device)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, entity.ErrDeviceNotFound
	}

	if err != nil {
		log.Error("failed to find device", zap.Error(err))
		return nil, err
	}

	return &device, nil
}

func (s *Service) Update(ctx context.Context, id string, input *gateway.DeviceUpdateInput) (*entity.Device, error) {
	log := logger.FromContext(ctx, s.log).With(
		zap.String("op", "mongo.Update"),
		zap.String("id", id),
	)
	filter := bson.M{"_id": id, "version": input.Version}
	update := bson.M{
		"$set": bson.M{
			"name":  input.Name,
			"brand": input.Brand,
			"state": *input.State},
		"$inc": bson.M{"version": 1}}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updated entity.Device
	err := s.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, s.disambiguateMiss(ctx, id)
	}
	if err != nil {
		log.Error("failed to update device", zap.Error(err))
		return nil, err
	}

	return &updated, nil
}

func (s *Service) Patch(ctx context.Context, id string, input *gateway.DevicePatchInput) (*entity.Device, error) {
	log := logger.FromContext(ctx, s.log).With(
		zap.String("op", "mongo.Patch"),
		zap.String("id", id),
	)
	set := bson.M{}
	if input.Name != nil {
		set["name"] = *input.Name
	}
	if input.Brand != nil {
		set["brand"] = *input.Brand
	}
	if input.State != nil {
		set["state"] = *input.State
	}

	update := bson.M{"$inc": bson.M{"version": 1}}
	if len(set) > 0 {
		update["$set"] = set
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updated entity.Device

	err := s.collection.FindOneAndUpdate(ctx, bson.M{"_id": id, "version": input.Version}, update, opts).Decode(&updated)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, s.disambiguateMiss(ctx, id)
	}

	if err != nil {
		log.Error("failed to patch device", zap.Error(err))
		return nil, err
	}

	return &updated, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	log := logger.FromContext(ctx, s.log).With(
		zap.String("op", "mongo.Delete"),
		zap.String("id", id),
	)
	res, err := s.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		log.Error("failed to delete device", zap.Error(err))
		return err
	}
	if res.DeletedCount == 0 {
		log.Debug("device not found")
		return entity.ErrDeviceNotFound
	}
	return nil
}

func (s *Service) List(ctx context.Context, filter *gateway.DeviceListFilter) (*entity.DevicePage, error) {
	log := logger.FromContext(ctx, s.log).With(
		zap.String("op", "mongo.List"),
	)

	log.Debug("filtering devices", zap.Any("filter", *filter))

	q := bson.M{}
	if filter.Brand != nil {
		q["brand"] = *filter.Brand
	}
	if filter.State != nil {
		q["state"] = *filter.State
	}
	if filter.Cursor != "" {
		q["_id"] = bson.M{"$gt": filter.Cursor}
	}

	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}).SetLimit(int64(filter.Limit) + 1)

	cursor, err := s.collection.Find(ctx, q, opts)
	if err != nil {
		log.Error("failed to list devices", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	items := make([]entity.Device, 0)
	if err := cursor.All(ctx, &items); err != nil {
		log.Error("failed to decode devices", zap.Error(err))
		return nil, err
	}

	next := ""
	if len(items) > filter.Limit {
		next = items[filter.Limit-1].ID
		items = items[:filter.Limit]
	}

	return &entity.DevicePage{
		Items:      items,
		NextCursor: next,
	}, nil
}
