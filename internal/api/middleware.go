package api

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"runtime/debug"
	"technical-challenge/internal/logger"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"
)

func Recovery(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error("recovered from panic",
						zap.Any("panic", err),
						zap.ByteString("stack", debug.Stack()))
					writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

const requestIDHeader = "X-Request-ID"

type requestIDCtxKey struct{}

func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(requestIDHeader)
			if id == "" {
				id = ulid.MustNew(ulid.Now(), rand.Reader).String()
			}
			w.Header().Set(requestIDHeader, id)
			ctx := context.WithValue(r.Context(), requestIDCtxKey{}, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDCtxKey{}).(string)
	return id
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

func RequestLogger(base *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			log := base.With(
				zap.String("request_id", RequestIDFromContext(r.Context())),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			)

			ctx := logger.WithContext(r.Context(), log)
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r.WithContext(ctx))
			log.Info("request completed",
				zap.Int("status", sw.status),
				zap.Duration("duration", time.Since(start)))
		})
	}
}

func MaxBytes(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}

func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func OpenAPIValidator(spec []byte) (func(http.Handler) http.Handler, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to load openapi spec: %w", err)
	}

	if err := doc.Validate(loader.Context); err != nil {
		return nil, fmt.Errorf("failed to validate openapi spec: %w", err)
	}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		return nil, fmt.Errorf("build openapi router: %w", err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route, pathParams, err := router.FindRoute(r)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			input := &openapi3filter.RequestValidationInput{
				Request:    r,
				Route:      route,
				PathParams: pathParams,
			}

			if err := openapi3filter.ValidateRequest(r.Context(), input); err != nil {
				writeError(w, http.StatusBadRequest, "validation_error", err.Error())
				return
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}

// Chaining made simple, ugly to read tho
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
