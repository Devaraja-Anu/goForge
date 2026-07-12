package api

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestPageParams(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantSize   int
		wantOffset int
	}{
		{"defaults with no params", "", defaultPageSize, 0},
		{"page 2 with default size", "page=2", defaultPageSize, defaultPageSize},
		{"custom size and page", "page_size=5&page=3", 5, 10},
		{"page_size below 1 falls back to default", "page_size=0", defaultPageSize, 0},
		{"page_size above max is capped", "page_size=1000", maxPageSize, 0},
		{"page below 1 falls back to 1", "page=-1", defaultPageSize, 0},
		{"non-numeric page_size falls back to default", "page_size=abc", defaultPageSize, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?"+tt.query, nil)
			gotSize, gotOffset := pageParams(r)
			if gotSize != tt.wantSize {
				t.Errorf("size = %d, want %d", gotSize, tt.wantSize)
			}
			if gotOffset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", gotOffset, tt.wantOffset)
			}
		})
	}
}

func TestQueryCSV(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{"missing key returns nil", "", nil},
		{"single value", "tags=a", []string{"a"}},
		{"multiple values", "tags=a,b,c", []string{"a", "b", "c"}},
		{"trims whitespace around values", "tags=%20a%20,%20b", []string{"a", "b"}},
		{"skips empty segments", "tags=a,,b", []string{"a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?"+tt.query, nil)
			got := queryCSV(r, "tags")
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestQueryInt(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		fallback int
		want     int
	}{
		{"missing key returns fallback", "", 42, 42},
		{"valid int returned", "n=7", 42, 7},
		{"non-numeric returns fallback", "n=abc", 42, 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?"+tt.query, nil)
			if got := QueryInt(r, "n", tt.fallback); got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestStrictQueryInt(t *testing.T) {
	t.Run("missing key errors", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		_, err := strictQueryInt(r, "n")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("non-numeric errors", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?n=abc", nil)
		_, err := strictQueryInt(r, "n")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("valid int returned with no error", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?n=9", nil)
		got, err := strictQueryInt(r, "n")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 9 {
			t.Errorf("got %d, want 9", got)
		}
	})
}

func TestPgUUIDConversions(t *testing.T) {
	id := uuid.New()

	t.Run("round trip through toPgUUID/fromPgUUID", func(t *testing.T) {
		pg := toPgUUID(id)
		if !pg.Valid {
			t.Fatal("expected Valid=true")
		}
		if fromPgUUID(pg) != id {
			t.Errorf("round trip mismatch: got %v, want %v", fromPgUUID(pg), id)
		}
	})

	t.Run("pointer variant nil input", func(t *testing.T) {
		pg := toPgUUIDPointer(nil)
		if pg.Valid {
			t.Error("expected Valid=false for nil input")
		}
	})

	t.Run("pointer variant non-nil input", func(t *testing.T) {
		pg := toPgUUIDPointer(&id)
		if !pg.Valid {
			t.Fatal("expected Valid=true")
		}
		if fromPgUUID(pg) != id {
			t.Errorf("mismatch: got %v, want %v", fromPgUUID(pg), id)
		}
	})

	t.Run("fromPgUUIDPointer invalid returns nil", func(t *testing.T) {
		pg := toPgUUIDPointer(nil)
		if got := fromPgUUIDPointer(pg); got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("fromPgUUIDPointer valid returns matching pointer", func(t *testing.T) {
		pg := toPgUUID(id)
		got := fromPgUUIDPointer(pg)
		if got == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *got != id {
			t.Errorf("got %v, want %v", *got, id)
		}
	})
}

func TestReadJSON(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	t.Run("valid single object decodes cleanly", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"alice"}`))
		w := httptest.NewRecorder()

		var dst payload
		if err := readJSON(w, r, &dst); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dst.Name != "alice" {
			t.Errorf("got %q, want %q", dst.Name, "alice")
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/", strings.NewReader(`{`))
		w := httptest.NewRecorder()

		var dst payload
		err := readJSON(w, r, &dst)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("unknown field rejected", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/", strings.NewReader(`{"nickname":"al"}`))
		w := httptest.NewRecorder()

		var dst payload
		err := readJSON(w, r, &dst)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "unknown field") {
			t.Errorf("error = %q, want it to mention unknown field", err.Error())
		}
	})

	t.Run("empty body rejected", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/", strings.NewReader(``))
		w := httptest.NewRecorder()

		var dst payload
		err := readJSON(w, r, &dst)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("multiple json values rejected", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"a"}{"name":"b"}`))
		w := httptest.NewRecorder()

		var dst payload
		err := readJSON(w, r, &dst)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "only one JSON value") {
			t.Errorf("error = %q, want it to mention single JSON value", err.Error())
		}
	})

	t.Run("body over max size rejected", func(t *testing.T) {
		oversized := bytes.Repeat([]byte("a"), maxBodyBytes+1)
		body := `{"name":"` + string(oversized) + `"}`

		r := httptest.NewRequest("POST", "/", strings.NewReader(body))
		w := httptest.NewRecorder()

		var dst payload
		err := readJSON(w, r, &dst)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
