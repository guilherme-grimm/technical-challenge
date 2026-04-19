package mongo_test

import (
	"context"
	"crypto/rand"
	"log"
	"os"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcmongo "github.com/testcontainers/testcontainers-go/modules/mongodb"
	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"

	"technical-challenge/internal/domain/entity"
	"technical-challenge/internal/domain/gateway"
	db "technical-challenge/internal/resource/database/mongo"
)

var sharedClient *mongodriver.Client

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcmongo.Run(ctx, "mongo:7")
	if err != nil {
		log.Fatalf("start mongo container: %v", err)
	}

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		log.Fatalf("connection string: %v", err)
	}

	cli, err := mongodriver.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		_ = container.Terminate(ctx)
		log.Fatalf("connect: %v", err)
	}

	if err := cli.Ping(ctx, nil); err != nil {
		_ = cli.Disconnect(ctx)
		_ = container.Terminate(ctx)
		log.Fatalf("ping: %v", err)
	}
	sharedClient = cli

	code := m.Run()

	_ = cli.Disconnect(ctx)
	_ = container.Terminate(ctx)
	os.Exit(code)
}

func newService(t *testing.T) *db.Service {
	t.Helper()
	dbName := "test_" + ulid.MustNew(ulid.Now(), rand.Reader).String()
	svc, err := db.New(sharedClient, dbName, "devices", zap.NewNop())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	t.Cleanup(func() {
		_ = sharedClient.Database(dbName).Drop(context.Background())
	})
	return svc
}

func newID() string {
	return ulid.MustNew(ulid.Now(), rand.Reader).String()
}

func makeDevice(id, name, brand string, state entity.State) *entity.Device {
	return &entity.Device{
		ID:        id,
		Name:      name,
		Brand:     brand,
		State:     state,
		CreatedAt: time.Now().UTC().Truncate(time.Millisecond),
		Version:   1,
	}
}
func TestService_Create(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, ctx context.Context, svc *db.Service) *entity.Device
		wantErr bool
	}{
		{
			name: "OK - Successfully creates device",
			setup: func(t *testing.T, ctx context.Context, svc *db.Service) *entity.Device {
				return makeDevice(newID(), "phone", "acme", entity.StateAvailable)
			},
			wantErr: false,
		},
		{
			name: "Error - Duplicate ID",
			setup: func(t *testing.T, ctx context.Context, svc *db.Service) *entity.Device {
				d := makeDevice(newID(), "phone", "acme", entity.StateAvailable)
				err := svc.Create(ctx, d)
				// Use require here: if setup fails, the whole test case is invalid
				require.NoError(t, err, "setup failed on first Create")
				return d
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newService(t)
			ctx := context.Background()

			d := tt.setup(t, ctx, svc)
			err := svc.Create(ctx, d)

			if tt.wantErr {
				require.Error(t, err)
				return // Stop here, no need to check DB state
			}
			require.NoError(t, err)

			got, err := svc.Get(ctx, d.ID)
			require.NoError(t, err, "Get after Create failed")

			// assert.Equal handles the deep comparison, including the CreatedAt time fields
			assert.Equal(t, d, got, "device in DB should match the created device")
		})
	}
}

// ---- Get ----

func TestService_Get(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, ctx context.Context, svc *db.Service) string
		expectedErr error
	}{
		{
			name: "OK - Successfully retrieves device",
			setup: func(t *testing.T, ctx context.Context, svc *db.Service) string {
				d := makeDevice(newID(), "phone", "acme", entity.StateAvailable)
				err := svc.Create(ctx, d)
				require.NoError(t, err, "setup failed on Create")
				return d.ID
			},
			expectedErr: nil,
		},
		{
			name: "Error - Device Not Found",
			setup: func(t *testing.T, ctx context.Context, svc *db.Service) string {
				return newID()
			},
			expectedErr: entity.ErrDeviceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newService(t)
			ctx := context.Background()

			targetID := tt.setup(t, ctx, svc)
			got, err := svc.Get(ctx, targetID)

			if tt.expectedErr != nil {
				// require.ErrorIs specifically checks the error chain
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)

			// Simple assert for the final condition
			assert.NotNil(t, got)
			assert.Equal(t, targetID, got.ID, "retrieved ID should match target ID")
		})
	}
}

// ---- Update ----

func TestService_Update(t *testing.T) {
	// Define state variable here so it can be referenced by pointers in the setup funcs
	stateInUse := entity.StateInUse

	tests := []struct {
		name string
		// setup returns the target ID and the input payload for the Update method
		setup       func(t *testing.T, ctx context.Context, svc *db.Service) (string, *gateway.DeviceUpdateInput)
		expectedErr error
		// validate allows us to run custom assertions on the returned device for successful updates
		validate func(t *testing.T, updated *entity.Device)
	}{
		{
			name: "OK - Successfully updates device",
			setup: func(t *testing.T, ctx context.Context, svc *db.Service) (string, *gateway.DeviceUpdateInput) {
				d := makeDevice(newID(), "phone", "acme", entity.StateAvailable)
				err := svc.Create(ctx, d)
				require.NoError(t, err, "setup failed on Create")

				input := &gateway.DeviceUpdateInput{
					Name:    "tablet",
					Brand:   "globex",
					State:   &stateInUse,
					Version: 1, // Matches current version
				}
				return d.ID, input
			},
			expectedErr: nil,
			validate: func(t *testing.T, updated *entity.Device) {
				assert.Equal(t, "tablet", updated.Name, "Name should be updated")
				assert.Equal(t, "globex", updated.Brand, "Brand should be updated")
				assert.Equal(t, entity.StateInUse, updated.State, "State should be updated")
				assert.Equal(t, int64(2), updated.Version, "Version should be incremented to 2")
			},
		},
		{
			name: "Error - Wrong Version (Conflict)",
			setup: func(t *testing.T, ctx context.Context, svc *db.Service) (string, *gateway.DeviceUpdateInput) {
				d := makeDevice(newID(), "phone", "acme", entity.StateAvailable)
				err := svc.Create(ctx, d)
				require.NoError(t, err, "setup failed on Create")

				input := &gateway.DeviceUpdateInput{
					Name:    "x",
					Brand:   "y",
					State:   &stateInUse,
					Version: 99, // Intentional mismatch
				}
				return d.ID, input
			},
			expectedErr: entity.ErrVersionConflict,
		},
		{
			name: "Error - Device Not Found",
			setup: func(t *testing.T, ctx context.Context, svc *db.Service) (string, *gateway.DeviceUpdateInput) {
				input := &gateway.DeviceUpdateInput{
					Name:    "x",
					Brand:   "y",
					State:   &stateInUse,
					Version: 1,
				}
				return newID(), input // ID doesn't exist in DB
			},
			expectedErr: entity.ErrDeviceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newService(t)
			ctx := context.Background()

			// Prepare state and get the inputs
			targetID, input := tt.setup(t, ctx, svc)

			// Execute the method under test
			updated, err := svc.Update(ctx, targetID, input)

			// Validate error expectations
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return // Stop execution, nothing more to check
			}
			require.NoError(t, err)
			require.NotNil(t, updated)

			// Run custom validation if provided
			if tt.validate != nil {
				tt.validate(t, updated)
			}
		})
	}
}

// ---- Patch ----

func TestService_Patch(t *testing.T) {
	// Define pointer variables here so they can be referenced easily in the setups
	newName := "tablet"

	tests := []struct {
		name string
		// setup returns the target ID and the input payload for the Patch method
		setup       func(t *testing.T, ctx context.Context, svc *db.Service) (string, *gateway.DevicePatchInput)
		expectedErr error
		// validate allows us to run custom assertions on the returned device
		validate func(t *testing.T, updated *entity.Device)
	}{
		{
			name: "OK - Partial update (Name only)",
			setup: func(t *testing.T, ctx context.Context, svc *db.Service) (string, *gateway.DevicePatchInput) {
				d := makeDevice(newID(), "phone", "acme", entity.StateAvailable)
				err := svc.Create(ctx, d)
				require.NoError(t, err, "setup failed on Create")

				input := &gateway.DevicePatchInput{
					Name:    &newName, // Only updating the name
					Version: 1,        // Matches current version
				}
				return d.ID, input
			},
			expectedErr: nil,
			validate: func(t *testing.T, updated *entity.Device) {
				// Assert the field that SHOULD change
				assert.Equal(t, newName, updated.Name, "Name should be updated")
				assert.Equal(t, int64(2), updated.Version, "Version should be incremented to 2")

				// Assert the fields that SHOULD NOT change
				assert.Equal(t, "acme", updated.Brand, "Brand should remain unchanged")
				assert.Equal(t, entity.StateAvailable, updated.State, "State should remain unchanged")
			},
		},
		{
			name: "Error - Wrong Version (Conflict)",
			setup: func(t *testing.T, ctx context.Context, svc *db.Service) (string, *gateway.DevicePatchInput) {
				d := makeDevice(newID(), "phone", "acme", entity.StateAvailable)
				err := svc.Create(ctx, d)
				require.NoError(t, err, "setup failed on Create")

				input := &gateway.DevicePatchInput{
					Name:    &newName,
					Version: 99, // Intentional mismatch
				}
				return d.ID, input
			},
			expectedErr: entity.ErrVersionConflict,
		},
		{
			name: "Error - Device Not Found",
			setup: func(t *testing.T, ctx context.Context, svc *db.Service) (string, *gateway.DevicePatchInput) {
				input := &gateway.DevicePatchInput{
					Name:    &newName,
					Version: 1,
				}
				return newID(), input // ID doesn't exist in DB
			},
			expectedErr: entity.ErrDeviceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newService(t)
			ctx := context.Background()

			targetID, input := tt.setup(t, ctx, svc)

			updated, err := svc.Patch(ctx, targetID, input)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, updated)

			if tt.validate != nil {
				tt.validate(t, updated)
			}
		})
	}
}

// ---- Delete ----

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name string
		// setup prepares the DB and returns the ID of the device to delete
		setup       func(t *testing.T, ctx context.Context, svc *db.Service) string
		expectedErr error
		// validate allows us to confirm the device was completely removed from the DB
		validate func(t *testing.T, ctx context.Context, svc *db.Service, targetID string)
	}{
		{
			name: "OK - Successfully deletes device",
			setup: func(t *testing.T, ctx context.Context, svc *db.Service) string {
				d := makeDevice(newID(), "phone", "acme", entity.StateAvailable)
				err := svc.Create(ctx, d)
				require.NoError(t, err, "setup failed on Create")
				return d.ID
			},
			expectedErr: nil,
			validate: func(t *testing.T, ctx context.Context, svc *db.Service, targetID string) {
				// Attempt to get the deleted device to ensure it's gone
				_, err := svc.Get(ctx, targetID)
				require.ErrorIs(t, err, entity.ErrDeviceNotFound, "device should be not found after deletion")
			},
		},
		{
			name: "Error - Device Not Found",
			setup: func(t *testing.T, ctx context.Context, svc *db.Service) string {
				return newID() // Generates an ID that doesn't exist in the DB
			},
			expectedErr: entity.ErrDeviceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newService(t)
			ctx := context.Background()

			// Prepare state and get the target ID
			targetID := tt.setup(t, ctx, svc)

			// Execute the method under test
			err := svc.Delete(ctx, targetID)

			// Validate error expectations
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return // Stop execution, nothing more to check
			}
			require.NoError(t, err)

			// Run custom database state validation if provided
			if tt.validate != nil {
				tt.validate(t, ctx, svc, targetID)
			}
		})
	}
}

// ---- List ----

func TestService_List(t *testing.T) {
	brandAcme := "acme"
	stateInUse := entity.StateInUse

	tests := []struct {
		name string
		// setup populates the database and returns the filter to be used in the initial List() call
		setup       func(t *testing.T, ctx context.Context, svc *db.Service) *gateway.DeviceListFilter
		expectedErr error
		// validate checks the resulting page, and can execute further queries if needed (like for pagination)
		validate func(t *testing.T, ctx context.Context, svc *db.Service, page *entity.DevicePage)
	}{
		{
			name: "OK - Empty list",
			setup: func(t *testing.T, ctx context.Context, svc *db.Service) *gateway.DeviceListFilter {
				return &gateway.DeviceListFilter{Limit: 10}
			},
			expectedErr: nil,
			validate: func(t *testing.T, ctx context.Context, svc *db.Service, page *entity.DevicePage) {
				assert.Empty(t, page.Items, "Items should be empty")
				assert.Empty(t, page.NextCursor, "NextCursor should be empty")
			},
		},
		{
			name: "OK - List all (fits in one page)",
			setup: func(t *testing.T, ctx context.Context, svc *db.Service) *gateway.DeviceListFilter {
				for range 3 {
					err := svc.Create(ctx, makeDevice(newID(), "p", "acme", entity.StateAvailable))
					require.NoError(t, err, "setup failed on Create")
				}
				return &gateway.DeviceListFilter{Limit: 10}
			},
			expectedErr: nil,
			validate: func(t *testing.T, ctx context.Context, svc *db.Service, page *entity.DevicePage) {
				assert.Len(t, page.Items, 3, "Should return exactly 3 items")
				assert.Empty(t, page.NextCursor, "NextCursor should be empty because all items fit")
			},
		},
		{
			name: "OK - Filter by Brand",
			setup: func(t *testing.T, ctx context.Context, svc *db.Service) *gateway.DeviceListFilter {
				err := svc.Create(ctx, makeDevice(newID(), "a", "acme", entity.StateAvailable))
				require.NoError(t, err)
				err = svc.Create(ctx, makeDevice(newID(), "b", "globex", entity.StateAvailable))
				require.NoError(t, err)

				return &gateway.DeviceListFilter{Brand: &brandAcme, Limit: 10}
			},
			expectedErr: nil,
			validate: func(t *testing.T, ctx context.Context, svc *db.Service, page *entity.DevicePage) {
				require.Len(t, page.Items, 1, "Should filter down to 1 item")
				assert.Equal(t, brandAcme, page.Items[0].Brand, "Returned item should match the filtered brand")
			},
		},
		{
			name: "OK - Filter by State",
			setup: func(t *testing.T, ctx context.Context, svc *db.Service) *gateway.DeviceListFilter {
				err := svc.Create(ctx, makeDevice(newID(), "a", "acme", entity.StateAvailable))
				require.NoError(t, err)
				err = svc.Create(ctx, makeDevice(newID(), "b", "acme", entity.StateInUse))
				require.NoError(t, err)

				return &gateway.DeviceListFilter{State: &stateInUse, Limit: 10}
			},
			expectedErr: nil,
			validate: func(t *testing.T, ctx context.Context, svc *db.Service, page *entity.DevicePage) {
				require.Len(t, page.Items, 1, "Should filter down to 1 item")
				assert.Equal(t, stateInUse, page.Items[0].State, "Returned item should match the filtered state")
			},
		},
		{
			name: "OK - Pagination logic",
			setup: func(t *testing.T, ctx context.Context, svc *db.Service) *gateway.DeviceListFilter {
				// 1ms gaps → distinct ULID timestamps → deterministic ascending _id sort.
				for i := range 3 {
					err := svc.Create(ctx, makeDevice(newID(), "p", "acme", entity.StateAvailable))
					require.NoError(t, err, "setup failed on Create %d", i)
					time.Sleep(time.Millisecond)
				}
				return &gateway.DeviceListFilter{Limit: 2}
			},
			expectedErr: nil,
			validate: func(t *testing.T, ctx context.Context, svc *db.Service, page1 *entity.DevicePage) {
				// Validate Page 1
				assert.Len(t, page1.Items, 2, "Page 1 should contain exactly 2 items")
				require.NotEmpty(t, page1.NextCursor, "Page 1 NextCursor should not be empty")

				// Fetch Page 2 using the cursor from Page 1
				page2, err := svc.List(ctx, &gateway.DeviceListFilter{Limit: 2, Cursor: page1.NextCursor})
				require.NoError(t, err, "fetching page 2 failed")

				// Validate Page 2
				assert.Len(t, page2.Items, 1, "Page 2 should contain the remaining 1 item")
				assert.Empty(t, page2.NextCursor, "Page 2 NextCursor should be empty (end of list)")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newService(t)
			ctx := context.Background()

			// Prepare state and get the filter
			filter := tt.setup(t, ctx, svc)

			// Execute the initial method under test
			page, err := svc.List(ctx, filter)

			// Validate error expectations
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return // Stop execution, nothing more to check
			}
			require.NoError(t, err)
			require.NotNil(t, page)

			// Run custom validations (including follow-up queries for pagination)
			if tt.validate != nil {
				tt.validate(t, ctx, svc, page)
			}
		})
	}
}
