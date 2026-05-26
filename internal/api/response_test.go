package api

import (
	"net/http"
	"testing"
)

func TestDefaultErrorCodeCoversAPIStatuses(t *testing.T) {
	tests := map[int]string{
		http.StatusBadRequest:            "BAD_REQUEST",
		http.StatusUnauthorized:          "UNAUTHORIZED",
		http.StatusForbidden:             "FORBIDDEN",
		http.StatusNotFound:              "NOT_FOUND",
		http.StatusMethodNotAllowed:      "METHOD_NOT_ALLOWED",
		http.StatusConflict:              "CONFLICT",
		http.StatusGone:                  "GONE",
		http.StatusRequestEntityTooLarge: "PAYLOAD_TOO_LARGE",
		http.StatusTooManyRequests:       "RATE_LIMITED",
		http.StatusBadGateway:            "UPSTREAM_ERROR",
		http.StatusServiceUnavailable:    "SERVICE_UNAVAILABLE",
		http.StatusInternalServerError:   "INTERNAL_ERROR",
	}

	for status, want := range tests {
		if got := defaultErrorCode(status, false); got != want {
			t.Fatalf("status %d error code = %q, want %q", status, got, want)
		}
	}
	if got := defaultErrorCode(http.StatusOK, true); got != "" {
		t.Fatalf("success error code = %q", got)
	}

	// 4xx 未命中 → REQUEST_FAILED；5xx 未命中 → INTERNAL_ERROR。
	// 这两条兜底是前端 errcode.ts 决定 toast 文案的关键 fallback，
	// 一旦被改成 ErrBadRequest / ErrUpstream 之类，所有 1xx-3xx-with-error
	// 的边角 case 都会被错误归类。
	if got := defaultErrorCode(http.StatusTeapot, false); got != "REQUEST_FAILED" {
		t.Fatalf("unmapped 4xx fallback = %q, want REQUEST_FAILED", got)
	}
	if got := defaultErrorCode(http.StatusNotImplemented, false); got != "INTERNAL_ERROR" {
		t.Fatalf("unmapped 5xx fallback = %q, want INTERNAL_ERROR", got)
	}
}
