package api

import "testing"

// ─── Test Battery ────────────────────────────────────────────────────────────

var ticketTests = []apiTest{}

// ─── Run Test ────────────────────────────────────────────────────────────────

func TestTickets(t *testing.T) {
	testRoutesInMemory(t, authTests)
	testRoutesHTTP(t, authTests)
}

// ─────────────────────────────────────────────────────────────────────────────
