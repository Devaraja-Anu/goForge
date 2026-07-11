package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/devaraja-anu/blueprint/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

const maxBodyBytes = 1_048_576 // 1MB

func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var (
			syntaxErr        *json.SyntaxError
			unmarshalTypeErr *json.UnmarshalTypeError
			invalidUnmarshal *json.InvalidUnmarshalError
		)

		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("malformed json at character %d", syntaxErr.Offset)
		// io.ErrUnexpectedEOF should be checked before io.EOF.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("malformed json: unexpected end of Input")
		case errors.As(err, &unmarshalTypeErr):
			if unmarshalTypeErr.Field != "" {
				return fmt.Errorf("incorrect type for field %q", unmarshalTypeErr.Field)
			}
			return fmt.Errorf("incorrect type at character %d", unmarshalTypeErr.Offset)
		//  this way to check EOF is correct for json.Decoder, not json.Unmarshal
		case errors.Is(err, io.EOF):
			return errors.New("request body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field"):
			field := strings.TrimPrefix(err.Error(), `json: unknown field `)
			return fmt.Errorf("unknown field %s", field)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("request body must not exceed %d bytes", maxBodyBytes)
		case errors.As(err, &invalidUnmarshal):
			panic(fmt.Sprintf("readJSON: invalid unmarshal target: %v", err))
		default:
			return err
		}

	}

	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain only one JSON value")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, val any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(val); err != nil {
		logger.FromContext(r.Context()).Error("failed to encode json response", "error", err)
	}
}

func urlParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func queryString(r *http.Request, key, fallback string) string {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	return v
}

func queryCSV(r *http.Request, key string) []string {

	v := r.URL.Query().Get(key)
	if v == "" {
		return nil
	}

	parts := strings.Split(v, ",")
	output := make([]string, 0, len(parts))

	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			output = append(output, t)
		}
	}

	return output
}

// for optional parameters
func QueryInt(r *http.Request, key string, fallback int) int {

	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}

	return n
}

// for strictly required parameters
func strictQueryInt(r *http.Request, key string) (int, error) {

	v := r.URL.Query().Get(key)
	if v == "" {
		return 0, fmt.Errorf("missing required query parameter: %q", key)
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("query parameter %q must be an integer", key)
	}

	return n, nil
}

func pageParams(r *http.Request) (int, int) {

	size := QueryInt(r, "page_size", defaultPageSize)
	page := QueryInt(r, "page", 1)

	if size < 1 {
		size = defaultPageSize
	} else if size > maxPageSize {
		size = maxPageSize
	}

	if page < 1 {
		page = 1
	}

	offset := (page - 1) * size
	return size, offset
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func toPgUUIDPointer(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func fromPgUUID(id pgtype.UUID) uuid.UUID {
	return uuid.UUID(id.Bytes)
}

func fromPgUUIDPointer(id pgtype.UUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}
	val := uuid.UUID(id.Bytes)
	return &val
}

func fromPgTimestamptz(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
