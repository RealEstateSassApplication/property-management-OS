package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type ReadinessChecker interface {
	Ping(context.Context) error
}

type OperationalConfig struct {
	MaxBodyBytes int64
	EnableHSTS   bool
}

type responseRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *responseRecorder) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseRecorder) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func NewOperationalHandler(next http.Handler, readiness ReadinessChecker, logger *slog.Logger, cfg OperationalConfig) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = 1 << 20
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		requestID := validRequestID(r.Header.Get("X-Request-ID"))
		if requestID == "" {
			requestID = newRequestID()
		}
		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
		w.Header().Set("Cache-Control", "no-store")
		if cfg.EnableHSTS {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		recorder := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("http request panicked", "requestId", requestID, "method", r.Method, "path", r.URL.Path, "panic", recovered)
				if !recorder.wroteHeader {
					writeError(recorder, http.StatusInternalServerError, "internal_error", "request could not be completed")
				}
			}
			logger.Info("http request", "requestId", requestID, "method", r.Method, "path", r.URL.Path, "status", recorder.status, "durationMs", time.Since(started).Milliseconds())
		}()

		if r.URL.Path == "/readyz" {
			if r.Method != http.MethodGet {
				recorder.Header().Set("Allow", http.MethodGet)
				writeError(recorder, http.StatusMethodNotAllowed, "method_not_allowed", "readiness endpoint only supports GET")
				return
			}
			if readiness == nil {
				writeError(recorder, http.StatusServiceUnavailable, "not_ready", "readiness checker is not configured")
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			err := readiness.Ping(ctx)
			cancel()
			if err != nil {
				logger.Warn("readiness check failed", "requestId", requestID, "error", err)
				writeError(recorder, http.StatusServiceUnavailable, "not_ready", "database is unavailable")
				return
			}
			writeJSON(recorder, http.StatusOK, map[string]string{"status": "ready"})
			return
		}

		if r.Body != nil && r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
			r.Body = http.MaxBytesReader(recorder, r.Body, cfg.MaxBodyBytes)
		}
		next.ServeHTTP(recorder, r)
	})
}

func validRequestID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 128 {
		return ""
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return ""
	}
	return value
}

func newRequestID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format("20060102150405.000000000")))
	}
	return hex.EncodeToString(bytes[:])
}
