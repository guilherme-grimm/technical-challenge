package api

import (
	"technical-challenge/internal/api/openapi"
	"technical-challenge/internal/domain/entity"
	"technical-challenge/internal/domain/gateway"
)

func toOpenAPIDevice(device *entity.Device) *openapi.Device {
	return &openapi.Device{
		Id:        device.ID,
		Name:      device.Name,
		Brand:     device.Brand,
		State:     openapi.State(device.State.String()),
		CreatedAt: device.CreatedAt,
		Version:   device.Version,
	}
}

func toOpenAPIPage(p *entity.DevicePage) openapi.DeviceListResponse {
	items := make([]openapi.Device, 0)
	for _, d := range p.Items {
		items = append(items, *toOpenAPIDevice(&d))
	}
	return openapi.DeviceListResponse{
		Items:      items,
		NextCursor: p.NextCursor,
	}
}

func toCreateInput(req openapi.CreateDeviceRequest) *gateway.DeviceCreateInput {
	in := &gateway.DeviceCreateInput{
		Name:  req.Name,
		Brand: req.Brand,
	}
	if req.State != nil {
		s := entity.State(*req.State)
		in.State = &s
	}
	return in
}

func toUpdateInput(req openapi.UpdateDeviceRequest) *gateway.DeviceUpdateInput {
	state := entity.State(req.State)
	return &gateway.DeviceUpdateInput{
		Name:    req.Name,
		Brand:   req.Brand,
		State:   &state,
		Version: req.Version,
	}
}

func toPatchInput(req openapi.PatchDeviceRequest) *gateway.DevicePatchInput {
	in := &gateway.DevicePatchInput{
		Name:    req.Name,
		Brand:   req.Brand,
		Version: req.Version,
	}
	if req.State != nil {
		s := entity.State(*req.State)
		in.State = &s
	}
	return in
}

func toListFilter(req openapi.ListDevicesParams) *gateway.DeviceListFilter {
	f := &gateway.DeviceListFilter{Brand: req.Brand, Limit: 50}
	if req.State != nil {
		s := entity.State(*req.State)
		f.State = &s
	}

	if req.Cursor != nil {
		f.Cursor = *req.Cursor
	}

	if req.Limit != nil {
		f.Limit = *req.Limit
	}

	return f
}
