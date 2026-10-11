package util

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var (
	ErrEmptyBody       = errors.New("body is empty")
	ErrBodyTooLarge    = errors.New("body too large")
	ErrInvalidJSON     = errors.New("body contains invalid JSON")
	ErrInvalidJSONType = errors.New("body contains invalid JSON type")
	ErrJSONValues      = errors.New("body contains more than one JSON value")
)

// Map wraps into a type the typical structure of JSON
type Map map[string]any

// ─── JSON Encoder/decoder ────────────────────────────────────────────────────

// EncodeJSON writes a JSON response with the given HTTP status code and headers.
func EncodeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data == nil {
		return nil
	}

	return json.NewEncoder(w).Encode(data)
}

// DecodeJSON reads JSON from the request body into a target struct, enforcing strict rules:
// - Max 1MB body limit to prevent OOM DOS vectors
// - Disallows unknown fields in payloads
// - Ensures single JSON object stream
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := int64(1_048_576) // 1MB limit
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return ErrInvalidJSON

		case errors.Is(err, io.ErrUnexpectedEOF):
			return ErrInvalidJSON

		case errors.As(err, &unmarshalTypeError):
			return ErrInvalidJSONType

		case errors.Is(err, io.EOF):
			return ErrEmptyBody

		case err.Error() == "http: request body too large":
			return ErrBodyTooLarge

		default:
			return err
		}
	}

	// Ensure there is only one JSON object in the stream
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return ErrJSONValues
	}

	return nil
}

// ─── Shortcut Methods ────────────────────────────────────────────────────────

// RespondError sends a standardized JSON error response
func RespondError(w http.ResponseWriter, status int, message string) {
	_ = EncodeJSON(w, status, Map{"error": message})
}

// ─────────────────────────────────────────────────────────────────────────────
