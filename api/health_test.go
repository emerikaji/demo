package api

import (
	"net/http"
	"testing"
)

// ─── Test Battery ────────────────────────────────────────────────────────────

var healthTests = []apiTest{
	{
		name:           "health check should return status available",
		route:          "/v1/health",
		method:         http.MethodGet,
		body:           nil,
		expectedStatus: http.StatusOK,
		expectedBody:   `{"status":"available"}`,
	},
}

// ─── Run Test ────────────────────────────────────────────────────────────────

func TestHealth(t *testing.T) {
	testRoutesInMemory(t, healthTests)
	testRoutesHTTP(t, healthTests)
}

// ─────────────────────────────────────────────────────────────────────────────
