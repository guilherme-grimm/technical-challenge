package api

import (
	"net/http"

	"go.uber.org/zap"

	"technical-challenge/internal/api/openapi"
	"technical-challenge/internal/domain/gateway"
)

type Handler struct {
	logger *zap.Logger
	svc    gateway.DeviceService
}

func NewHandler(logger *zap.Logger) *Handler {
	return &Handler{logger: logger}
}

var _ openapi.ServerInterface = (*Handler)(nil)

func (h *Handler) CreateDevice(w http.ResponseWriter, r *http.Request) {
	// TODO: decode openapi.CreateDeviceJSONRequestBody, call the service,
	// map errors via writeDomainError, write 201 + Location
	writeError(w, http.StatusNotImplemented, "not_implemented", "createDevice is not implemented")
}

func (h *Handler) ListDevices(w http.ResponseWriter, r *http.Request, params openapi.ListDevicesParams) {
	// TODO: translate params into device.Filter, call the service, marshal
	// into openapi.DeviceListResponse
	writeError(w, http.StatusNotImplemented, "not_implemented", "listDevices is not implemented")
}

func (h *Handler) GetDevice(w http.ResponseWriter, r *http.Request, id openapi.DeviceIDPath) {
	// TODO: call the service, map ErrDeviceNotFound → 404
	writeError(w, http.StatusNotImplemented, "not_implemented", "getDevice is not implemented")
}

func (h *Handler) UpdateDevice(w http.ResponseWriter, r *http.Request, id openapi.DeviceIDPath) {
	// TODO: decode UpdateDeviceJSONRequestBody, call the service, map
	// ErrVersionConflict → 409, ErrDeviceNotFound → 404
	writeError(w, http.StatusNotImplemented, "not_implemented", "updateDevice is not implemented")
}

func (h *Handler) PatchDevice(w http.ResponseWriter, r *http.Request, id openapi.DeviceIDPath) {
	// TODO: decode PatchDeviceJSONRequestBody, call the service
	writeError(w, http.StatusNotImplemented, "not_implemented", "patchDevice is not implemented")
}

func (h *Handler) DeleteDevice(w http.ResponseWriter, r *http.Request, id openapi.DeviceIDPath) {
	// TODO: call the service, map ErrDeviceInUse → 409
	writeError(w, http.StatusNotImplemented, "not_implemented", "deleteDevice is not implemented")
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// TODO: ping Mongo and report degraded on failure
	writeJSON(w, http.StatusOK, openapi.HealthResponse{
		Status: openapi.Ok,
	})
}

