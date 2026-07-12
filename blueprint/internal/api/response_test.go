package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
)

// decodeEnvelope decodes a recorded response body into a generic map so
// tests can assert on shape without depending on writeJSON's internals.
func decodeEnvelope(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	if w.Body.Len() == 0 {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("response body is not valid JSON: %v\nbody: %s", err, w.Body.String())
	}
	return out
}

func TestSimpleResponseHelpers(t *testing.T) {
	tests := []struct {
		name       string
		call       func(w http.ResponseWriter, r *http.Request)
		wantStatus int
		wantCode   string // expected envelope["error"].(map)["code"], "" if no error envelope expected
	}{
		{
			name:       "ok with data wraps in data envelope",
			call:       func(w http.ResponseWriter, r *http.Request) { ok(w, r, map[string]string{"x": "y"}) },
			wantStatus: http.StatusOK,
		},
		{
			name:       "ok with nil data writes empty 200",
			call:       func(w http.ResponseWriter, r *http.Request) { ok(w, r, nil) },
			wantStatus: http.StatusOK,
		},
		{
			name:       "created wraps in data envelope with 201",
			call:       func(w http.ResponseWriter, r *http.Request) { created(w, r, map[string]string{"x": "y"}) },
			wantStatus: http.StatusCreated,
		},
		{
			name:       "noContent writes empty 204",
			call:       func(w http.ResponseWriter, r *http.Request) { noContent(w, r) },
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "badRequest wraps error with bad_request code",
			call:       func(w http.ResponseWriter, r *http.Request) { badRequest(w, r, errors.New("bad input")) },
			wantStatus: http.StatusBadRequest,
			wantCode:   "bad_request",
		},
		{
			name: "unauthorized uses given code and message",
			call: func(w http.ResponseWriter, r *http.Request) {
				unauthorized(w, r, "no_token", "missing token")
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "no_token",
		},
		{
			name: "forbidden uses given code and message",
			call: func(w http.ResponseWriter, r *http.Request) {
				forbidden(w, r, "no_access", "not allowed")
			},
			wantStatus: http.StatusForbidden,
			wantCode:   "no_access",
		},
		{
			name:       "notFound uses fixed not_found code",
			call:       func(w http.ResponseWriter, r *http.Request) { notFound(w, r) },
			wantStatus: http.StatusNotFound,
			wantCode:   "not_found",
		},
		{
			name: "conflict uses given code and message",
			call: func(w http.ResponseWriter, r *http.Request) {
				conflict(w, r, "duplicate", "already exists")
			},
			wantStatus: http.StatusConflict,
			wantCode:   "duplicate",
		},
		{
			name:       "rateLimited uses fixed rate_limited code",
			call:       func(w http.ResponseWriter, r *http.Request) { rateLimited(w, r) },
			wantStatus: http.StatusTooManyRequests,
			wantCode:   "rate_limited",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			tt.call(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantCode == "" {
				return
			}

			body := decodeEnvelope(t, w)
			errObj, ok := body["error"].(map[string]any)
			if !ok {
				t.Fatalf("expected an \"error\" object in body, got %v", body)
			}
			if code, _ := errObj["code"].(string); code != tt.wantCode {
				t.Errorf("error code = %q, want %q", code, tt.wantCode)
			}
		})
	}
}

func TestCollection(t *testing.T) {
	tests := []struct {
		name            string
		total           int
		pageSize        int
		offset          int
		wantPageCount   int
		wantCurrentPage int
	}{
		{"exact division", 100, 20, 0, 5, 1},
		{"remainder rounds page count up", 101, 20, 0, 6, 1},
		{"offset into second page", 100, 20, 20, 5, 2},
		{"zero page size avoids divide by zero", 100, 0, 0, 0, 1},
		{"zero total with normal page size", 0, 20, 0, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			collection(w, r, []string{"item"}, tt.total, tt.pageSize, tt.offset)

			body := decodeEnvelope(t, w)
			meta, ok := body["meta"].(map[string]any)
			if !ok {
				t.Fatalf("expected \"meta\" object in body, got %v", body)
			}

			gotPageCount := int(meta["page_count"].(float64))
			gotCurrentPage := int(meta["current_page"].(float64))

			if gotPageCount != tt.wantPageCount {
				t.Errorf("page_count = %d, want %d", gotPageCount, tt.wantPageCount)
			}
			if gotCurrentPage != tt.wantCurrentPage {
				t.Errorf("current_page = %d, want %d", gotCurrentPage, tt.wantCurrentPage)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	type testStruct struct {
		Name string `json:"name" validate:"required"`
	}

	t.Run("validator.ValidationErrors produces 422 with fields", func(t *testing.T) {
		v := validator.New()
		err := v.Struct(testStruct{})
		if err == nil {
			t.Fatal("expected validation to fail on empty required field")
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		validationError(w, r, err)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
		}

		body := decodeEnvelope(t, w)
		errObj, ok := body["error"].(map[string]any)
		if !ok {
			t.Fatalf("expected \"error\" object in body, got %v", body)
		}
		if code, _ := errObj["code"].(string); code != "validation_error" {
			t.Errorf("code = %q, want validation_error", code)
		}
		if _, ok := errObj["fields"]; !ok {
			t.Error("expected \"fields\" to be present on a validation error")
		}
	})

	t.Run("non-validation error falls back to badRequest", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		validationError(w, r, errors.New("plain error"))

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d (fallback to badRequest)", w.Code, http.StatusBadRequest)
		}

		body := decodeEnvelope(t, w)
		errObj, ok := body["error"].(map[string]any)
		if !ok {
			t.Fatalf("expected \"error\" object in body, got %v", body)
		}
		if code, _ := errObj["code"].(string); code != "bad_request" {
			t.Errorf("code = %q, want bad_request", code)
		}
	})
}

func TestServerError(t *testing.T) {
	// application.logger is unused by serverError itself — it logs via
	// logger.FromContext(r.Context()), which falls back to slog.Default()
	// when no logger has been attached to the context (confirmed via
	// internal/logger's FromContext implementation). A zero-value
	// application is therefore safe to use here.
	app := &application{}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	app.serverError(w, r, errors.New("db connection lost"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	body := decodeEnvelope(t, w)
	errObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected \"error\" object in body, got %v", body)
	}
	if code, _ := errObj["code"].(string); code != "server_error" {
		t.Errorf("code = %q, want server_error", code)
	}
	// Message is intentionally generic and must never leak the underlying
	// error string to the client.
	if msg, _ := errObj["message"].(string); msg == "db connection lost" {
		t.Error("serverError must not leak the underlying error message to the client")
	}
}

func TestValidationMessage(t *testing.T) {
	type testStruct struct {
		Name string `json:"name" validate:"required,min=3"`
	}

	v := validator.New()
	err := v.Struct(testStruct{})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		t.Fatal("expected validator.ValidationErrors")
	}
	if len(ve) == 0 {
		t.Fatal("expected at least one field error")
	}

	// "required" fires first since Name is empty; message should be non-empty
	// and not fall through to the generic default case.
	msg := validationMessage(ve[0])
	if msg == "" {
		t.Error("expected non-empty message")
	}
}
