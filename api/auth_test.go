package api

import (
	"net/http"
	"testing"
)

// ─── Test Battery ────────────────────────────────────────────────────────────

var authTests = []apiTest{

	// ─── Register ──────────────────────────────────────────────────────────────

	{
		name:           "register - empty body should return 400",
		route:          "/v1/auth/register",
		method:         http.MethodPost,
		body:           []byte(``),
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"body must not be empty"}`,
	},
	{
		name:           "register - missing required fields should return 400",
		route:          "/v1/auth/register",
		method:         http.MethodPost,
		body:           []byte(`{"email":"janedoe@example.com"}`),
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"name, email, password, and role are required"}`,
	},
	{
		name:           "register - invalid role should return 400",
		route:          "/v1/auth/register",
		method:         http.MethodPost,
		body:           []byte(`{"name":"Jane","email":"janedoe@example.com","password":"password123","role":"admin"}`),
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"invalid role: must be 'organizer' or 'customer'"}`,
	},
	{
		name:           "register - successful customer creation should return 201",
		route:          "/v1/auth/register",
		method:         http.MethodPost,
		body:           []byte(`{"name":"Jane","email":"janedoe@example.com","password":"password123","role":"customer"}`),
		expectedStatus: http.StatusCreated,
		expectedBody:   `{"message":"user registered successfully"}`,
	},
	{
		name:           "register - duplicate email should return 409",
		route:          "/v1/auth/register",
		method:         http.MethodPost,
		body:           []byte(`{"name":"Jane Doe 2","email":"janedoe@example.com","password":"password123","role":"customer"}`),
		expectedStatus: http.StatusConflict,
		expectedBody:   `{"error":"user with this email already exists"}`,
	},

	// ─── Login ─────────────────────────────────────────────────────────────────

	{
		name:           "login - empty body should return 400",
		route:          "/v1/auth/login",
		method:         http.MethodPost,
		body:           []byte(``),
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"body must not be empty"}`,
	},
	{
		name:           "login - missing credentials should return 400",
		route:          "/v1/auth/login",
		method:         http.MethodPost,
		body:           []byte(`{"email":"janedoe@example.com"}`),
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"email and password are required"}`,
	},
	{
		name:           "login - non-existent email should return 401",
		route:          "/v1/auth/login",
		method:         http.MethodPost,
		body:           []byte(`{"email":"alexandrina@example.com","password":"password123"}`),
		expectedStatus: http.StatusUnauthorized,
		expectedBody:   `{"error":"invalid credentials"}`,
	},
	{
		name:           "login - wrong password should return 401",
		route:          "/v1/auth/login",
		method:         http.MethodPost,
		body:           []byte(`{"email":"janedoe@example.com","password":"password321"}`),
		expectedStatus: http.StatusUnauthorized,
		expectedBody:   `{"error":"invalid credentials"}`,
	},
	{
		name:           "login - successful login should return 200",
		route:          "/v1/auth/login",
		method:         http.MethodPost,
		body:           []byte(`{"email":"janedoe@example.com","password":"password123"}`),
		expectedStatus: http.StatusOK,
		expectedBody:   `{"message":"login successful"}`,
	},
}

// ─── Run Test ────────────────────────────────────────────────────────────────

func TestAuth(t *testing.T) {
	testRoutesInMemory(t, authTests)
	testRoutesHTTP(t, authTests)
}

// ─────────────────────────────────────────────────────────────────────────────
