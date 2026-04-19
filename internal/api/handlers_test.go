package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"technical-challenge/internal/api"
	"technical-challenge/internal/api/openapi"
	"technical-challenge/internal/resource/database/memory"
	"technical-challenge/internal/service/device"
)

type stubPinger struct{ err error }

func (s stubPinger) Ping(ctx context.Context) error { return s.err }

func setupHandler(t *testing.T, pinger api.Pinger) http.Handler {
	t.Helper()
	db := memory.NewService()
	svc, err := device.New(db, zap.NewNop())
	require.NoError(t, err)
	h := api.NewHandler(zap.NewNop(), svc, pinger)
	return openapi.HandlerFromMux(h, http.NewServeMux())
}

func do(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body != nil {
		buf, err := json.Marshal(body)
		require.NoError(t, err)
		r = httptest.NewRequest(method, path, bytes.NewReader(buf))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func decode[T any](t *testing.T, body io.Reader) T {
	t.Helper()
	var v T
	require.NoError(t, json.NewDecoder(body).Decode(&v))
	return v
}

func TestCreateDevice(t *testing.T) {
	h := setupHandler(t, stubPinger{})

	w := do(t, h, http.MethodPost, "/devices", openapi.CreateDeviceRequest{
		Name: "phone", Brand: "acme",
	})

	require.Equal(t, http.StatusCreated, w.Code)
	assert.NotEmpty(t, w.Header().Get("Location"))

	got := decode[openapi.Device](t, w.Body)
	assert.NotEmpty(t, got.Id)
	assert.Equal(t, "phone", got.Name)
	assert.Equal(t, openapi.StateAvailable, got.State)
	assert.Equal(t, int64(1), got.Version)
}

func TestCreateDevice_InvalidJSON(t *testing.T) {
	h := setupHandler(t, stubPinger{})

	r := httptest.NewRequest(http.MethodPost, "/devices", bytes.NewReader([]byte("{not json")))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateDevice_MissingName(t *testing.T) {
	h := setupHandler(t, stubPinger{})

	w := do(t, h, http.MethodPost, "/devices", openapi.CreateDeviceRequest{Brand: "acme"})

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	err := decode[openapi.ErrorResponse](t, w.Body)
	assert.Equal(t, "validation_error", err.Code)
}

func TestGetDevice_NotFound(t *testing.T) {
	h := setupHandler(t, stubPinger{})

	w := do(t, h, http.MethodGet, "/devices/missing-id", nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	err := decode[openapi.ErrorResponse](t, w.Body)
	assert.Equal(t, "not_found", err.Code)
}

func TestGetDevice_OK(t *testing.T) {
	h := setupHandler(t, stubPinger{})

	created := decode[openapi.Device](t, do(t, h, http.MethodPost, "/devices",
		openapi.CreateDeviceRequest{Name: "phone", Brand: "acme"}).Body)

	w := do(t, h, http.MethodGet, "/devices/"+created.Id, nil)
	require.Equal(t, http.StatusOK, w.Code)

	got := decode[openapi.Device](t, w.Body)
	assert.Equal(t, created.Id, got.Id)
}

func TestUpdateDevice_VersionConflict(t *testing.T) {
	h := setupHandler(t, stubPinger{})

	created := decode[openapi.Device](t, do(t, h, http.MethodPost, "/devices",
		openapi.CreateDeviceRequest{Name: "phone", Brand: "acme"}).Body)

	w := do(t, h, http.MethodPut, "/devices/"+created.Id, openapi.UpdateDeviceRequest{
		Name: "tablet", Brand: "globex", State: openapi.StateAvailable, Version: 99,
	})

	require.Equal(t, http.StatusConflict, w.Code)
	err := decode[openapi.ErrorResponse](t, w.Body)
	assert.Equal(t, "version_conflict", err.Code)
}

func TestUpdateDevice_NameChangeBlockedWhileInUse(t *testing.T) {
	h := setupHandler(t, stubPinger{})

	inUse := openapi.StateInUse
	created := decode[openapi.Device](t, do(t, h, http.MethodPost, "/devices",
		openapi.CreateDeviceRequest{Name: "phone", Brand: "acme", State: &inUse}).Body)

	w := do(t, h, http.MethodPut, "/devices/"+created.Id, openapi.UpdateDeviceRequest{
		Name: "tablet", Brand: "acme", State: openapi.StateInUse, Version: created.Version,
	})

	require.Equal(t, http.StatusConflict, w.Code)
	err := decode[openapi.ErrorResponse](t, w.Body)
	assert.Equal(t, "device_in_use_immutable", err.Code)
}

func TestPatchDevice_BrandChangeBlockedWhileInUse(t *testing.T) {
	h := setupHandler(t, stubPinger{})

	inUse := openapi.StateInUse
	created := decode[openapi.Device](t, do(t, h, http.MethodPost, "/devices",
		openapi.CreateDeviceRequest{Name: "phone", Brand: "acme", State: &inUse}).Body)

	newBrand := "globex"
	w := do(t, h, http.MethodPatch, "/devices/"+created.Id, openapi.PatchDeviceRequest{
		Brand: &newBrand, Version: created.Version,
	})

	require.Equal(t, http.StatusConflict, w.Code)
	err := decode[openapi.ErrorResponse](t, w.Body)
	assert.Equal(t, "device_in_use_immutable", err.Code)
}

func TestPatchDevice_StateOnlyAllowedWhileInUse(t *testing.T) {
	h := setupHandler(t, stubPinger{})

	inUse := openapi.StateInUse
	created := decode[openapi.Device](t, do(t, h, http.MethodPost, "/devices",
		openapi.CreateDeviceRequest{Name: "phone", Brand: "acme", State: &inUse}).Body)

	avail := openapi.StateAvailable
	w := do(t, h, http.MethodPatch, "/devices/"+created.Id, openapi.PatchDeviceRequest{
		State: &avail, Version: created.Version,
	})

	require.Equal(t, http.StatusOK, w.Code)
	got := decode[openapi.Device](t, w.Body)
	assert.Equal(t, openapi.StateAvailable, got.State)
}

func TestDeleteDevice_InUse(t *testing.T) {
	h := setupHandler(t, stubPinger{})

	inUse := openapi.StateInUse
	created := decode[openapi.Device](t, do(t, h, http.MethodPost, "/devices",
		openapi.CreateDeviceRequest{Name: "phone", Brand: "acme", State: &inUse}).Body)

	w := do(t, h, http.MethodDelete, "/devices/"+created.Id, nil)

	require.Equal(t, http.StatusConflict, w.Code)
	err := decode[openapi.ErrorResponse](t, w.Body)
	assert.Equal(t, "device_in_use", err.Code)
}

func TestDeleteDevice_OK(t *testing.T) {
	h := setupHandler(t, stubPinger{})

	created := decode[openapi.Device](t, do(t, h, http.MethodPost, "/devices",
		openapi.CreateDeviceRequest{Name: "phone", Brand: "acme"}).Body)

	w := do(t, h, http.MethodDelete, "/devices/"+created.Id, nil)
	require.Equal(t, http.StatusNoContent, w.Code)

	w = do(t, h, http.MethodGet, "/devices/"+created.Id, nil)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteDevice_NotFound(t *testing.T) {
	h := setupHandler(t, stubPinger{})

	w := do(t, h, http.MethodDelete, "/devices/missing", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListDevices_FilterByBrand(t *testing.T) {
	h := setupHandler(t, stubPinger{})

	_ = do(t, h, http.MethodPost, "/devices", openapi.CreateDeviceRequest{Name: "a", Brand: "acme"})
	_ = do(t, h, http.MethodPost, "/devices", openapi.CreateDeviceRequest{Name: "b", Brand: "globex"})

	w := do(t, h, http.MethodGet, "/devices?brand=acme", nil)
	require.Equal(t, http.StatusOK, w.Code)

	page := decode[openapi.DeviceListResponse](t, w.Body)
	require.Len(t, page.Items, 1)
	assert.Equal(t, "acme", page.Items[0].Brand)
}

func TestHealthCheck_OK(t *testing.T) {
	h := setupHandler(t, stubPinger{err: nil})

	w := do(t, h, http.MethodGet, "/healthz", nil)
	require.Equal(t, http.StatusOK, w.Code)

	resp := decode[openapi.HealthResponse](t, w.Body)
	assert.Equal(t, openapi.Ok, resp.Status)
}

func TestHealthCheck_Degraded(t *testing.T) {
	h := setupHandler(t, stubPinger{err: errors.New("unreachable")})

	w := do(t, h, http.MethodGet, "/healthz", nil)
	require.Equal(t, http.StatusServiceUnavailable, w.Code)

	resp := decode[openapi.HealthResponse](t, w.Body)
	assert.Equal(t, openapi.Degraded, resp.Status)
	require.NotNil(t, resp.Details)
	assert.Contains(t, (*resp.Details)["database"], "unreachable")
}

func TestMethodNotAllowed_OptionsHandled(t *testing.T) {
	h := setupHandler(t, stubPinger{})

	r := httptest.NewRequest(http.MethodOptions, "/devices", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	// openapi.HandlerFromMux responds 405 for unsupported methods on a known path;
	// CORS preflight is handled by the CORS middleware, not covered here.
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
