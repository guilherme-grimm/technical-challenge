package main

import (
	"context"
	"technical-challenge/internal/domain/gateway"
	"technical-challenge/internal/resource/database"
	db "technical-challenge/internal/resource/database/mongo"
	"technical-challenge/internal/service/device"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.uber.org/zap"
)

func setupDatabase(ctx context.Context, logger *zap.Logger, uri, dbName, dbCollection string) (database.Service, error) {
	cli, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := cli.Ping(ctx, readpref.Primary()); err != nil {
		_ = cli.Disconnect(ctx)
		return nil, err
	}

	svc, err := db.New(cli, dbName, dbCollection, logger)
	if err != nil {
		_ = cli.Disconnect(ctx)
		return nil, err
	}

	return svc, nil
}

func setupService(logger *zap.Logger, db database.Service) (gateway.DeviceService, error) {
	svc, err := device.New(db, logger)
	if err != nil {
		return nil, err
	}
	return svc, nil
}
