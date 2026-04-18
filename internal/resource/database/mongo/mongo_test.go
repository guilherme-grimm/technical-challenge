package mongo_test

import (
	"context"
	"technical-challenge/internal/domain/entity"
	"technical-challenge/internal/domain/gateway"
	db "technical-challenge/internal/resource/database/mongo"
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

func TestService_Create(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		client     *mongo.Client
		database   string
		collection string
		log        *zap.Logger
		// Named input parameters for target function.
		device  *entity.Device
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := db.New(tt.client, tt.database, tt.collection, tt.log)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			gotErr := s.Create(context.Background(), tt.device)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Create() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Create() succeeded unexpectedly")
			}
		})
	}
}

func TestService_List(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		client     *mongo.Client
		database   string
		collection string
		log        *zap.Logger
		// Named input parameters for target function.
		filter  *gateway.DeviceListFilter
		want    *entity.DevicePage
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := db.New(tt.client, tt.database, tt.collection, tt.log)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got, gotErr := s.List(context.Background(), tt.filter)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("List() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("List() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_Get(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		client     *mongo.Client
		database   string
		collection string
		log        *zap.Logger
		// Named input parameters for target function.
		id      string
		want    *entity.Device
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := db.New(tt.client, tt.database, tt.collection, tt.log)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got, gotErr := s.Get(context.Background(), tt.id)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Get() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Get() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		client     *mongo.Client
		database   string
		collection string
		log        *zap.Logger
		// Named input parameters for target function.
		id      string
		input   *gateway.DeviceUpdateInput
		want    *entity.Device
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := db.New(tt.client, tt.database, tt.collection, tt.log)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got, gotErr := s.Update(context.Background(), tt.id, tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Update() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Update() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("Update() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_Patch(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		client     *mongo.Client
		database   string
		collection string
		log        *zap.Logger
		// Named input parameters for target function.
		id      string
		input   *gateway.DevicePatchInput
		want    *entity.Device
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := db.New(tt.client, tt.database, tt.collection, tt.log)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got, gotErr := s.Patch(context.Background(), tt.id, tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Patch() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Patch() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("Patch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		client     *mongo.Client
		database   string
		collection string
		log        *zap.Logger
		// Named input parameters for target function.
		id      string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := db.New(tt.client, tt.database, tt.collection, tt.log)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			gotErr := s.Delete(context.Background(), tt.id)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Delete() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Delete() succeeded unexpectedly")
			}
		})
	}
}
