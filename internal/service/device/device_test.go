package device_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"technical-challenge/internal/domain/entity"
	"technical-challenge/internal/domain/gateway"
	"technical-challenge/internal/resource/database"
	"technical-challenge/internal/resource/database/memory"
	"technical-challenge/internal/service/device"
)

// setupTest creates a fresh in-memory DB and the gateway service for each test.
func setupTest(t *testing.T) (*device.Service, database.Service) {
	db := memory.NewService()
	svc, err := device.New(db, zap.NewNop())
	require.NoError(t, err)
	return svc, db
}

func TestNew(t *testing.T) {
	db := memory.NewService()
	log := zap.NewNop()

	_, err := device.New(nil, log)
	assert.ErrorIs(t, err, entity.ErrEmptyClient)

	_, err = device.New(db, nil)
	assert.ErrorIs(t, err, entity.ErrEmptyLogger)
}
func TestService_Create(t *testing.T) {
	stateAvailable := entity.StateAvailable
	stateInvalid := entity.State("invalid")

	tests := []struct {
		name        string
		input       *gateway.DeviceCreateInput
		expectedErr error
	}{
		{
			name: "OK - Creates device with default state if omitted",
			input: &gateway.DeviceCreateInput{
				Name:  "phone",
				Brand: "acme",
			},
			expectedErr: nil,
		},
		{
			name: "OK - Creates device with explicit state",
			input: &gateway.DeviceCreateInput{
				Name:  "tablet",
				Brand: "globex",
				State: &stateAvailable,
			},
			expectedErr: nil,
		},
		{
			name:        "Error - Nil Input",
			input:       nil,
			expectedErr: entity.ErrEmptyDeviceCreateInput,
		},
		{
			name: "Error - Empty Name",
			input: &gateway.DeviceCreateInput{
				Name:  "",
				Brand: "acme",
			},
			expectedErr: entity.ErrEmptyDeviceName,
		},
		{
			name: "Error - Empty Brand",
			input: &gateway.DeviceCreateInput{
				Name:  "phone",
				Brand: "",
			},
			expectedErr: entity.ErrEmptyDeviceBrand,
		},
		{
			name: "Error - Invalid State",
			input: &gateway.DeviceCreateInput{
				Name:  "phone",
				Brand: "acme",
				State: &stateInvalid,
			},
			expectedErr: entity.ErrInvalidDeviceState,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := setupTest(t)
			ctx := context.Background()

			got, err := svc.Create(ctx, tt.input)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)

			// Dynamic assertions for fields generated inside the black box
			assert.NotEmpty(t, got.ID, "ID should be generated")
			assert.NotZero(t, got.CreatedAt, "CreatedAt should be generated")
			assert.Equal(t, int64(1), got.Version, "Version should initialize at ")

			// Check input mapping
			assert.Equal(t, tt.input.Name, got.Name)
			assert.Equal(t, tt.input.Brand, got.Brand)

			if tt.input.State != nil {
				assert.Equal(t, *tt.input.State, got.State)
			} else {
				assert.Equal(t, entity.StateAvailable, got.State, "Should default to Available")
			}
		})
	}
}
func TestService_Delete(t *testing.T) {
	tests := []struct {
		name string
		// setup populates the DB and returns the ID to attempt to delete
		setup       func(t *testing.T, ctx context.Context, db database.Service) string
		expectedErr error
	}{
		{
			name: "OK - Successfully deletes an available device",
			setup: func(t *testing.T, ctx context.Context, db database.Service) string {
				d := &entity.Device{
					ID:        "dev_123",
					Name:      "phone",
					Brand:     "acme",
					State:     entity.StateAvailable, // Safe to delete
					CreatedAt: time.Now(),
					Version:   1,
				}
				require.NoError(t, db.Create(ctx, d))
				return d.ID
			},
			expectedErr: nil,
		},
		{
			name: "OK - Successfully deletes an inactive device",
			setup: func(t *testing.T, ctx context.Context, db database.Service) string {
				d := &entity.Device{
					ID:    "dev_456",
					State: entity.StateInactive, // Safe to delete
				}
				require.NoError(t, db.Create(ctx, d))
				return d.ID
			},
			expectedErr: nil,
		},
		{
			name: "Error - Device is In-Use (Business Logic)",
			setup: func(t *testing.T, ctx context.Context, db database.Service) string {
				d := &entity.Device{
					ID:    "dev_789",
					State: entity.StateInUse, // Blocks deletion
				}
				require.NoError(t, db.Create(ctx, d))
				return d.ID
			},
			expectedErr: entity.ErrDeviceInUse,
		},
		{
			name: "Error - Empty ID provided",
			setup: func(t *testing.T, ctx context.Context, db database.Service) string {
				return ""
			},
			expectedErr: entity.ErrEmptyDeviceID,
		},
		{
			name: "Error - Device Not Found",
			setup: func(t *testing.T, ctx context.Context, db database.Service) string {
				return "non_existent_id"
			},
			expectedErr: entity.ErrDeviceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, dbFake := setupTest(t)
			ctx := context.Background()

			targetID := tt.setup(t, ctx, dbFake)

			err := svc.Delete(ctx, targetID)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)

			_, err = dbFake.Get(ctx, targetID)
			require.ErrorIs(t, err, entity.ErrDeviceNotFound, "Device should be removed from DB")
		})
	}
}

func TestService_Get(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T, ctx context.Context, db database.Service) string
		expectedErr error
	}{
		{
			name: "OK - Retrieves existing device",
			setup: func(t *testing.T, ctx context.Context, db database.Service) string {
				d := &entity.Device{
					ID:    "dev_123",
					Name:  "phone",
					Brand: "acme",
					State: entity.StateAvailable,
				}
				require.NoError(t, db.Create(ctx, d))
				return d.ID
			},
			expectedErr: nil,
		},
		{
			name: "Error - Empty ID provided",
			setup: func(t *testing.T, ctx context.Context, db database.Service) string {
				return ""
			},
			expectedErr: entity.ErrEmptyDeviceID,
		},
		{
			name: "Error - Device Not Found",
			setup: func(t *testing.T, ctx context.Context, db database.Service) string {
				return "non_existent_id"
			},
			expectedErr: entity.ErrDeviceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, dbFake := setupTest(t)
			ctx := context.Background()

			targetID := tt.setup(t, ctx, dbFake)
			got, err := svc.Get(ctx, targetID)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, targetID, got.ID)
		})
	}
}
func TestService_Update(t *testing.T) {
	t.Parallel()

	stateInUse := entity.StateInUse
	stateInvalid := entity.State("invalid")

	tests := []struct {
		name        string
		setup       func(t *testing.T, ctx context.Context, db database.Service) (string, *gateway.DeviceUpdateInput)
		expectedErr error
	}{
		{
			name: "OK - Successfully updates device",
			setup: func(t *testing.T, ctx context.Context, db database.Service) (string, *gateway.DeviceUpdateInput) {
				d := &entity.Device{ID: "dev_123", Name: "old", Brand: "old", State: entity.StateAvailable, Version: 1}
				require.NoError(t, db.Create(ctx, d))

				return d.ID, &gateway.DeviceUpdateInput{
					Name:    "tablet",
					Brand:   "globex",
					State:   &stateInUse,
					Version: 1,
				}
			},
			expectedErr: nil,
		},
		{
			name: "Error - Version Conflict (Stale Update)",
			setup: func(t *testing.T, ctx context.Context, db database.Service) (string, *gateway.DeviceUpdateInput) {
				d := &entity.Device{ID: "dev_456", Name: "phone", Brand: "acme", State: entity.StateAvailable, Version: 2}
				require.NoError(t, db.Create(ctx, d))

				return d.ID, &gateway.DeviceUpdateInput{
					Name:    "tablet",
					Brand:   "globex",
					State:   &stateInUse,
					Version: 1,
				}
			},
			expectedErr: entity.ErrVersionConflict,
		},
		{
			name: "Error - Empty ID",
			setup: func(t *testing.T, ctx context.Context, db database.Service) (string, *gateway.DeviceUpdateInput) {
				return "", &gateway.DeviceUpdateInput{Name: "tablet", Brand: "globex", State: &stateInUse, Version: 1}
			},
			expectedErr: entity.ErrEmptyDeviceID,
		},
		{
			name: "Error - Nil Input",
			setup: func(t *testing.T, ctx context.Context, db database.Service) (string, *gateway.DeviceUpdateInput) {
				return "dev_123", nil
			},
			expectedErr: entity.ErrEmptyDeviceUpdateInput,
		},
		{
			name: "Error - Invalid State",
			setup: func(t *testing.T, ctx context.Context, db database.Service) (string, *gateway.DeviceUpdateInput) {
				return "dev_123", &gateway.DeviceUpdateInput{Name: "tablet", Brand: "globex", State: &stateInvalid, Version: 1}
			},
			expectedErr: entity.ErrInvalidDeviceState,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, dbFake := setupTest(t)
			ctx := context.Background()

			targetID, input := tt.setup(t, ctx, dbFake)
			got, err := svc.Update(ctx, targetID, input)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, input.Name, got.Name)
			assert.Equal(t, *input.State, got.State)
			assert.Equal(t, int64(2), got.Version, "Version should be incremented")
		})
	}
}
func TestService_List(t *testing.T) {
	t.Parallel()

	brandAcme := "acme"

	tests := []struct {
		name        string
		setup       func(t *testing.T, ctx context.Context, db database.Service) *gateway.DeviceListFilter
		expectedErr error
		validate    func(t *testing.T, ctx context.Context, svc gateway.DeviceService, page1 *entity.DevicePage)
	}{
		{
			name: "Error - Nil Filter",
			setup: func(t *testing.T, ctx context.Context, db database.Service) *gateway.DeviceListFilter {
				return nil
			},
			expectedErr: entity.ErrEmptyListFilter,
		},
		{
			name: "OK - List respects filter",
			setup: func(t *testing.T, ctx context.Context, db database.Service) *gateway.DeviceListFilter {
				require.NoError(t, db.Create(ctx, &entity.Device{ID: "1", Brand: "acme"}))
				require.NoError(t, db.Create(ctx, &entity.Device{ID: "2", Brand: "globex"}))

				return &gateway.DeviceListFilter{Brand: &brandAcme, Limit: 10}
			},
			expectedErr: nil,
			validate: func(t *testing.T, ctx context.Context, svc gateway.DeviceService, page *entity.DevicePage) {
				require.Len(t, page.Items, 1)
				assert.Equal(t, brandAcme, page.Items[0].Brand)
			},
		},
		{
			name: "OK - Pagination limits and cursors work",
			setup: func(t *testing.T, ctx context.Context, db database.Service) *gateway.DeviceListFilter {
				// insert 3 items but limit the query to 2
				require.NoError(t, db.Create(ctx, &entity.Device{ID: "1", Brand: "acme"}))
				require.NoError(t, db.Create(ctx, &entity.Device{ID: "2", Brand: "acme"}))
				require.NoError(t, db.Create(ctx, &entity.Device{ID: "3", Brand: "acme"}))

				return &gateway.DeviceListFilter{Limit: 2}
			},
			expectedErr: nil,
			validate: func(t *testing.T, ctx context.Context, svc gateway.DeviceService, page1 *entity.DevicePage) {
				assert.Len(t, page1.Items, 2, "Page 1 should hit the limit")
				require.NotEmpty(t, page1.NextCursor, "Page 1 should have a cursor for the next page")

				// second query using the cursor to get the remaining items
				page2, err := svc.List(ctx, &gateway.DeviceListFilter{Limit: 2, Cursor: page1.NextCursor})
				require.NoError(t, err)

				assert.Len(t, page2.Items, 1, "Page 2 should have the final item")
				assert.Empty(t, page2.NextCursor, "Page 2 should be the end of the list")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, dbFake := setupTest(t)
			ctx := context.Background()

			filter := tt.setup(t, ctx, dbFake)
			page, err := svc.List(ctx, filter)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, ctx, svc, page)
			}
		})
	}
}
