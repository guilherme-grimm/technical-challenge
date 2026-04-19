package memory

import (
	"context"
	"sort"
	"sync"

	"technical-challenge/internal/domain/entity"
	"technical-challenge/internal/domain/gateway"
	"technical-challenge/internal/resource/database"
)

var _ database.Service = (*Service)(nil)

type Service struct {
	mu      sync.RWMutex
	devices map[string]*entity.Device
}

func NewService() *Service {
	return &Service{
		devices: make(map[string]*entity.Device),
	}
}

func (s *Service) Create(ctx context.Context, device *entity.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.devices[device.ID]; exists {
		return entity.ErrDuplicateID
	}

	// Store a clone to prevent tests from mutating the DB state directly
	clone := *device
	s.devices[device.ID] = &clone
	return nil
}

func (s *Service) Get(ctx context.Context, id string) (*entity.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	device, exists := s.devices[id]
	if !exists {
		return nil, entity.ErrDeviceNotFound
	}

	clone := *device
	return &clone, nil
}

func (s *Service) Update(ctx context.Context, id string, input *gateway.DeviceUpdateInput) (*entity.Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	device, exists := s.devices[id]
	if !exists {
		return nil, entity.ErrDeviceNotFound
	}

	if device.Version != input.Version {
		return nil, entity.ErrVersionConflict
	}

	clone := *device
	clone.Name = input.Name
	clone.Brand = input.Brand
	if input.State != nil {
		clone.State = *input.State
	}
	clone.Version++

	s.devices[id] = &clone

	ret := clone
	return &ret, nil
}

func (s *Service) Patch(ctx context.Context, id string, input *gateway.DevicePatchInput) (*entity.Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	device, exists := s.devices[id]
	if !exists {
		return nil, entity.ErrDeviceNotFound
	}

	if device.Version != input.Version {
		return nil, entity.ErrVersionConflict
	}

	clone := *device
	if input.Name != nil {
		clone.Name = *input.Name
	}
	if input.Brand != nil {
		clone.Brand = *input.Brand
	}
	if input.State != nil {
		clone.State = *input.State
	}
	clone.Version++

	s.devices[id] = &clone

	ret := clone
	return &ret, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.devices[id]; !exists {
		return entity.ErrDeviceNotFound
	}

	delete(s.devices, id)
	return nil
}

func (s *Service) List(ctx context.Context, filter *gateway.DeviceListFilter) (*entity.DevicePage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]entity.Device, 0)

	for _, d := range s.devices {
		if filter != nil {
			if filter.Brand != nil && d.Brand != *filter.Brand {
				continue
			}
			if filter.State != nil && d.State != *filter.State {
				continue
			}
		}
		clone := *d
		items = append(items, clone)
	}

	// sort by ID to guarantee deterministic pagination (simulating Mongo's _id/ULID sort)
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	// apply Pagination (Cursor and Limit)
	startIndex := 0
	if filter != nil && filter.Cursor != "" {
		for i, d := range items {
			if d.ID == filter.Cursor {
				startIndex = i + 1
				break
			}
		}
	}

	limit := 10
	if filter != nil && filter.Limit > 0 {
		limit = filter.Limit
	}

	endIndex := startIndex + limit
	var nextCursor string

	if endIndex < len(items) {
		nextCursor = items[endIndex-1].ID
	} else {
		endIndex = len(items)
	}

	// Handle case where startIndex is completely out of bounds
	if startIndex >= len(items) {
		return &entity.DevicePage{Items: []entity.Device{}, NextCursor: ""}, nil
	}

	pageItems := items[startIndex:endIndex]

	return &entity.DevicePage{
		Items:      pageItems,
		NextCursor: nextCursor,
	}, nil
}

func (s *Service) Ping(ctx context.Context) error {
	return nil // Always healthy
}

func (s *Service) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices = make(map[string]*entity.Device) // Clear out on close
	return nil
}
