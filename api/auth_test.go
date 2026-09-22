package api

import (
	"context"
	"demo/domain"
	"demo/store"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
		expectedBody:   `{"error":"email, password, and role are required"}`,
	},
	{
		name:           "register - invalid role should return 400",
		route:          "/v1/auth/register",
		method:         http.MethodPost,
		body:           []byte(`{"email":"janedoe@example.com","password":"password123","role":"superuser"}`),
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"invalid role: must be 'organizer' or 'attendee'"}`,
	},
	{
		name:           "register - successful attendee creation should return 201",
		route:          "/v1/auth/register",
		method:         http.MethodPost,
		body:           []byte(`{"email":"janedoe@example.com","password":"password123","role":"attendee"}`),
		expectedStatus: http.StatusCreated,
		expectedBody:   `{"message":"user registered successfully"}`,
	},
	{
		name:           "register - duplicate email should return 409",
		route:          "/v1/auth/register",
		method:         http.MethodPost,
		body:           []byte(`{"email":"janedoe@example.com","password":"password123","role":"attendee"}`),
		expectedStatus: http.StatusConflict,
		expectedBody:   `{"error":"user with this email already exists"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			err := s.CreateUser(context.Background(), &domain.User{
				ID:           uuid.New(),
				Email:        "janedoe@example.com",
				PasswordHash: string(hash),
				Role:         domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}
		},
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
		name:           "login - non-existent user should return 401",
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
		body:           []byte(`{"email":"janedoe@example.com","password":"wrongpassword"}`),
		expectedStatus: http.StatusUnauthorized,
		expectedBody:   `{"error":"invalid credentials"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			err := s.CreateUser(context.Background(), &domain.User{
				ID:           uuid.New(),
				Email:        "janedoe@example.com",
				PasswordHash: string(hash),
				Role:         domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}
		},
	},
	{
		name:           "login - successful login should return 200 with JWT token",
		route:          "/v1/auth/login",
		method:         http.MethodPost,
		body:           []byte(`{"email":"janedoe@example.com","password":"password123"}`),
		expectedStatus: http.StatusOK,
		expectedBody:   `{"token":"a valid token"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			err := s.CreateUser(context.Background(), &domain.User{
				ID:           uuid.New(),
				Email:        "janedoe@example.com",
				PasswordHash: string(hash),
				Role:         domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}

			j.returnToken = "a valid token"
		},
	},
}

// ─── Run Test ────────────────────────────────────────────────────────────────

func TestAuth(t *testing.T) {
	testRoutesInMemory(t, authTests)
	testRoutesHTTP(t, authTests)
}

// ─────────────────────────────────────────────────────────────────────────────
