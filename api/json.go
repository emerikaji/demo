package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Map wraps into a type the typical structure of JSON
type Map map[string]any

// encodeJSON writes a JSON response with the given HTTP status code and headers.
func encodeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data == nil {
		return nil
	}

	return json.NewEncoder(w).Encode(data)
}

// decodeJSON reads JSON from the request body into a target struct, enforcing strict rules:
// - Max 1MB body limit to prevent OOM DOS vectors
// - Disallows unknown fields in payloads
// - Ensures single JSON object stream
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := int64(1_048_576) // 1MB limit
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not exceed %d bytes", maxBytes)

		default:
			return err
		}
	}

	// Ensure there is only one JSON object in the stream
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// respondError sends a standardized JSON error response
func respondError(w http.ResponseWriter, status int, message string) {
	_ = encodeJSON(w, status, Map{"error": message})
}
