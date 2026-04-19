package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"technical-challenge/internal/api/openapi"
	"technical-challenge/internal/domain/entity"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, openapi.ErrorResponse{
		Code:    code,
		Message: message,
	})
}

func writeDomainError(w http.ResponseWriter, log *zap.Logger, err error) {
	switch {
	case errors.Is(err, entity.ErrDeviceNotFound):
		writeError(w, http.StatusNotFound, "not_found", err.Error())

	case errors.Is(err, entity.ErrVersionConflict):
		writeError(w, http.StatusConflict, "version_conflict", err.Error())
	case errors.Is(err, entity.ErrDeviceInUse):
		writeError(w, http.StatusConflict, "device_in_use", err.Error())
	case errors.Is(err, entity.ErrDeviceInUseImmutable):
		writeError(w, http.StatusConflict, "device_in_use_immutable", err.Error())
	case errors.Is(err, entity.ErrDuplicateID):
		writeError(w, http.StatusConflict, "duplicate_id", err.Error())

	case errors.Is(err, entity.ErrInvalidDeviceState),
		errors.Is(err, entity.ErrEmptyDeviceName),
		errors.Is(err, entity.ErrEmptyDeviceBrand),
		errors.Is(err, entity.ErrEmptyDeviceID),
		errors.Is(err, entity.ErrEmptyDeviceCreateInput),
		errors.Is(err, entity.ErrEmptyDeviceUpdateInput),
		errors.Is(err, entity.ErrEmptyDevicePatchInput),
		errors.Is(err, entity.ErrEmptyListFilter):
		writeError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())

	default:
		log.Error("unexpected error", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
