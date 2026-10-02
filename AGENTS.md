# AGENTS.md

Guidance and conventions for AI agents and automated contributors working on this codebase.

---

## 1. Project Overview

This repository is a containerized Go demonstration API modeling a high-concurrency event ticketing system with user authentication, event management, and ticket reservation lifecycles.

- **Language & Runtime:** Go 1.27 (`go.mod` specifies `go 1.27.1`)
- **Architecture:** Clean, layered architecture with decoupled domain, storage, server, API routing, and utilities.
- **Dependencies:** Minimal external dependencies (standard library preferred):
  - `github.com/golang-jwt/jwt/v5`: JWT authentication
  - `github.com/google/uuid`: Unique identifier generation
  - `golang.org/x/crypto/bcrypt`: Password hashing
  - `golang.org/x/sync/errgroup`: Goroutine lifecycle coordination
- **Deployment:** Multi-stage Docker build (`scratch` base) deployed behind a Traefik reverse proxy with automated GitHub Actions CI/CD.

---

## 2. Architecture & Directory Layout

The codebase strictly enforces separation of concerns across packages:

```
├── api/             # HTTP route handlers, request validation, and API integration tests
│   ├── api.go       # Handler struct & route registration (Go 1.22+ ServeMux routing)
│   ├── api_test.go  # Test runner harness (in-memory & HTTP loopback test runners)
│   ├── auth.go      # /v1/auth routes (register, login)
│   ├── event.go     # /v1/events routes
│   ├── health.go    # /v1/health route
│   ├── ticket.go    # /v1/tickets routes
│   └── user.go      # /v1/user routes
├── domain/          # Pure business domain layer (zero HTTP or external storage dependencies)
│   ├── entities.go  # Core structs (User, Event, Ticket) and role/status enums
│   └── store.go     # Repository interfaces, filter types, and sentinel errors
├── server/          # HTTP server configuration and graceful shutdown mechanics
│   └── server.go    # Server wrapper, timeouts, and signal-aware Start()
├── store/           # Datastore implementations
│   └── memory.go    # Thread-safe in-memory store (sync.RWMutex, defensive copy)
├── util/            # Shared cross-cutting utilities
│   ├── json.go      # Strict JSON encoding/decoding and error response helpers
│   └── jwt.go       # JWT provider interface and HMAC implementation
├── .github/         # CI/CD workflows (test.yml, deploy.yml)
├── CHANGELOG.md     # Changelog following Keep a Changelog format
├── Dockerfile       # Multi-stage scratch Docker container
└── main.go          # Application composition root and graceful termination handling
```

---

## 3. Code Style & Design Conventions

### Section Headers & Box Drawing Comments
All Go source files group logical sections with Unicode box-drawing horizontal lines (`─`, U+2500):
```go
// ─── Section Title ───────────────────────────────────────────────────────────

// Code goes here...

// ─────────────────────────────────────────────────────────────────────────────
```
Maintain this exact visual pattern when adding new files, types, or functional groups.

### Line Endings & Formatting
- **Line Endings:** Always enforce **LF (`\n`)** line endings on all files.
- **Formatting:** Code must adhere strictly to standard `gofmt` / `goimports`.

### Error Handling & Responses
- **Domain Sentinel Errors:** Define new sentinel errors in [domain/store.go](file:///E:/Code/demo/domain/store.go) prefixed with `Err...` (e.g. `ErrNotFound`, `ErrEmailAlreadyExists`). Check them using `errors.Is(err, domain.Err...)`.
- **API Error Responses:** Never write ad-hoc JSON errors. Always use `util.RespondError(w, httpStatus, message)` to enforce the standard JSON error schema:
  ```json
  {"error": "error message here"}
  ```
- **Internal Errors:** Do not leak database or internal implementation details in 500 error responses; return generic messages like `"internal server error"`.

### Request & Response Processing
- Use `util.DecodeJSON(w, r, &req)` which automatically:
  - Enforces a 1MB payload size limit (`http.MaxBytesReader`).
  - Calls `dec.DisallowUnknownFields()` to prevent unknown or malformed payload keys.
  - Ensures a single valid JSON value in the request stream.
- Use `util.EncodeJSON(w, httpStatus, data)` or `util.Map` for structured JSON output.
- Normalize input strings where applicable:
  ```go
  req.Email = strings.TrimSpace(strings.ToLower(req.Email))
  req.Password = strings.TrimSpace(req.Password)
  ```

### Structs & JSON Field Naming
- Field names in JSON payloads use `snake_case` (e.g., `organizer_id`, `remaining_tickets`, `starts_at`).
- Sensitive or internal fields must be omitted from JSON serialization with `json:"-"` (e.g., `PasswordHash`, `ExpiresAt`).

### Concurrency & In-Memory Storage
- When reading or updating state in `MemoryStore`:
  - Acquire read lock (`s.mu.RLock()`) for queries; acquire write lock (`s.mu.Lock()`) for mutations.
  - Return **copies** of pointers or dereferenced structs so callers cannot cause race conditions on stored memory.
  - Verify compile-time interface adherence:
    ```go
    var (
        _ domain.UserRepository   = (*MemoryStore)(nil)
        _ domain.EventRepository  = (*MemoryStore)(nil)
        _ domain.TicketRepository = (*MemoryStore)(nil)
    )
    ```

---

## 4. Production Logic: Test-Driven Development (TDD)

The codebase strictly adheres to **Test-Driven Development (TDD)** for all feature development, route additions, and bug fixes:

1. **Tests Are Created First:**
   - Before implementing any production handler, business logic, or datastore method, write the test battery first.
   - Define test cases covering input validation, missing parameters, unauthorized/forbidden access, error responses, edge cases, and expected happy paths.
   - Route tests must be drafted in the corresponding `*_test.go` file (e.g. `api/event_test.go`, `api/ticket_test.go`) utilizing the `apiTest` structure before filling in handlers.
2. **Implementation Follows Tests:**
   - Once the tests specify the expected behavior, status codes, and JSON response bodies, write the production code to satisfy those tests.
3. **Verify with Race Detection:**
   - Run `go test -v -race ./...` to verify that all tests pass cleanly with zero race conditions or regressions.

---

## 5. Testing Conventions & Harness

The repository maintains strict test coverage executed via:
```bash
go test -v -race ./...
```

### Table-Driven Test Battery (`apiTest`)
All API route test suites in `api/` must follow the established `apiTest` structure defined in [api/api_test.go](file:///E:/Code/demo/api/api_test.go):

```go
type apiTest struct {
    name           string
    route          string
    method         string
    headers        map[string]string
    body           []byte
    expectedStatus int
    expectedBody   string
    setup          func(t *testing.T, s *store.MemoryStore, j *MockJWT)
}
```

### Dual Test Execution
Every route test battery must execute through both test harnesses:
```go
func TestFeature(t *testing.T) {
    testRoutesInMemory(t, featureTests) // Direct ServeHTTP with httptest.NewRecorder
    testRoutesHTTP(t, featureTests)     // Full loopback server with httptest.NewTestServer
}
```

- Enable `t.Parallel()` inside table-test iterations.
- Use `MockJWT` to mock authentication claims and tokens without making cryptographic calls.
- Use the `setup` closure to seed required entities in `store.MemoryStore`.

---

## 6. Branching & Git Workflow

- **Branch Structure:**
  - `dev`: Primary integration branch for active development and continuous testing.
  - `main`: Production release branch. Merging to `main` triggers automated build and deployment via GHCR and Traefik.
- **Changelog:**
  - Update [CHANGELOG.md](file:///E:/Code/demo/CHANGELOG.md) for notable changes following [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) standards under `## [Unreleased]`.
  - Maintain standard change categories: `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, `Security`.
- **Commit Messages:**
  - Concise, descriptive, lowercase imperative statements (e.g. `implement ticket reservation route. add tests.`).
