package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"technical-challenge/internal/api/openapi"
	"technical-challenge/internal/domain/gateway"
)

type Handler struct {
	logger *zap.Logger
	svc    gateway.DeviceService
}

func NewHandler(logger *zap.Logger, svc gateway.DeviceService) *Handler {
	return &Handler{
		logger: logger,
		svc:    svc,
	}
}

var _ openapi.ServerInterface = (*Handler)(nil)

func (h *Handler) CreateDevice(w http.ResponseWriter, r *http.Request) {
	var body openapi.CreateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	d, err := h.svc.Create(r.Context(), toCreateInput(body))
	if err != nil {
		writeDomainError(w, h.logger, err)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("/devices/%s", d.ID))
	writeJSON(w, http.StatusCreated, toOpenAPIDevice(d))
}

func (h *Handler) ListDevices(w http.ResponseWriter, r *http.Request, params openapi.ListDevicesParams) {
	page, err := h.svc.List(r.Context(), toListFilter(params))
	if err != nil {
		writeDomainError(w, h.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, toOpenAPIPage(page))
}

func (h *Handler) GetDevice(w http.ResponseWriter, r *http.Request, id openapi.DeviceIDPath) {
	d, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeDomainError(w, h.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, toOpenAPIDevice(d))
}

func (h *Handler) UpdateDevice(w http.ResponseWriter, r *http.Request, id openapi.DeviceIDPath) {
	var body openapi.UpdateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	d, err := h.svc.Update(r.Context(), id, toUpdateInput(body))
	if err != nil {
		writeDomainError(w, h.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, toOpenAPIDevice(d))
}

func (h *Handler) PatchDevice(w http.ResponseWriter, r *http.Request, id openapi.DeviceIDPath) {
	var body openapi.PatchDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	d, err := h.svc.Patch(r.Context(), id, toPatchInput(body))
	if err != nil {
		writeDomainError(w, h.logger, err)
		return
	}

	writeJSON(w, http.StatusOK, toOpenAPIDevice(d))
}

func (h *Handler) DeleteDevice(w http.ResponseWriter, r *http.Request, id openapi.DeviceIDPath) {
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeDomainError(w, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// TODO: ping Mongo and report degraded on failure
	writeJSON(w, http.StatusOK, openapi.HealthResponse{
		Status: openapi.Ok,
	})
}
