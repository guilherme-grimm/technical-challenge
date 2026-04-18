package mongo

import (
	"context"
	"errors"
	"technical-challenge/internal/domain/entity"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func (s *Service) disambiguateMiss(ctx context.Context, id string) error {
	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return entity.ErrDeviceNotFound
	}
	if err != nil {
		return err
	}
	return entity.ErrVersionConflict
}
