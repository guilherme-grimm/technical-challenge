package database

import (
	"context"
	"technical-challenge/internal/domain/entity"
)

// database contract, bound to domain interfaces

// TODO: properly define database operations
// in case of a multi type database, where there are no only devices, but more entries for the same ops, would go with generics
// for now, will be only for devices tho
type Service interface {
	Create(ctx context.Context, device entity.Device) error
	List(ctx context.Context) ([]entity.Device, error) // define if it'll ge one or many here
	Get(ctx context.Context) (*entity.Device, error)
	Update(ctx context.Context) error
	Patch(ctx context.Context) error
	Delete(ctx context.Context) error
}
