package api

import (
	"testing"
)

// ─── Test Battery ────────────────────────────────────────────────────────────

var eventTests = []apiTest{
	{
		name:  "",
		route: "/v1/events",
	},
}

// ─── Run Test ────────────────────────────────────────────────────────────────

func TestEvent(t *testing.T) {
	testRoutesInMemory(t, authTests)
	testRoutesHTTP(t, authTests)
}

// ─────────────────────────────────────────────────────────────────────────────
