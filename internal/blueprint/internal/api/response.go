package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/devaraja-anu/blueprint/internal/logger"
	"github.com/go-playground/validator/v10"
)

// envelope is the top-level JSON wrapper for all responses.
type envelope map[string]any

// errorPayload is the shape of every error response body.
type errorPayload struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"` // only present on validation errors
}

// --- success helpers ---

func ok(w http.ResponseWriter, r *http.Request, data any) {
	if data == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	writeJSON(w, r, http.StatusOK, envelope{"data": data})
}

func created(w http.ResponseWriter, r *http.Request, data any) {
	writeJSON(w, r, http.StatusCreated, envelope{"data": data})
}

func noContent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func collection(w http.ResponseWriter, r *http.Request, data any, total, pageSize, offset int) {
	pageCount := 0
	currentPage := 1
	if pageSize > 0 {
		pageCount = (total + pageSize - 1) / pageSize
		currentPage = (offset / pageSize) + 1
	}
	writeJSON(w, r, http.StatusOK, envelope{
		"data": data,
		"meta": map[string]int{
			"total":        total,
			"page_size":    pageSize,
			"offset":       offset,
			"page_count":   pageCount,
			"current_page": currentPage,
		},
	})
}

// --- error helpers ---

func badRequest(w http.ResponseWriter, r *http.Request, err error) {
	writeJSON(w, r, http.StatusBadRequest, envelope{
		"error": errorPayload{
			Code:    "bad_request",
			Message: err.Error(),
		},
	})
}

func validationError(w http.ResponseWriter, r *http.Request, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		fields := make(map[string]string, len(ve))
		for _, e := range ve {
			fields[e.Field()] = validationMessage(e)
		}
		writeJSON(w, r, http.StatusUnprocessableEntity, envelope{
			"error": errorPayload{
				Code:    "validation_error",
				Message: "One or more fields failed validation",
				Fields:  fields,
			},
		})
		return
	}
	badRequest(w, r, err)
}

func unauthorized(w http.ResponseWriter, r *http.Request, code, message string) {
	writeJSON(w, r, http.StatusUnauthorized, envelope{
		"error": errorPayload{
			Code:    code,
			Message: message,
		},
	})
}

func forbidden(w http.ResponseWriter, r *http.Request, code, message string) {
	writeJSON(w, r, http.StatusForbidden, envelope{
		"error": errorPayload{
			Code:    code,
			Message: message,
		},
	})
}

func notFound(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusNotFound, envelope{
		"error": errorPayload{
			Code:    "not_found",
			Message: "The requested resource could not be found",
		},
	})
}

func conflict(w http.ResponseWriter, r *http.Request, code, message string) {
	writeJSON(w, r, http.StatusConflict, envelope{
		"error": errorPayload{
			Code:    code,
			Message: message,
		},
	})
}

func rateLimited(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusTooManyRequests, envelope{
		"error": errorPayload{
			Code:    "rate_limited",
			Message: "Too many requests, please try again later",
		},
	})
}

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {

	logger.FromContext(r.Context()).Error("internal server error", "error", err, "path", r.URL.Path)

	writeJSON(w, r, http.StatusInternalServerError, envelope{
		"error": errorPayload{
			Code:    "server_error",
			Message: "An unexpected error occurred",
		},
	})
}

func validationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return fmt.Sprintf("Must be at least %s characters", e.Param())
	case "max":
		return fmt.Sprintf("Must be no more than %s characters", e.Param())
	case "len":
		return fmt.Sprintf("Must be exactly %s characters", e.Param())
	case "numeric":
		return "Must contain only numbers"
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", strings.ReplaceAll(e.Param(), " ", ", "))
	case "latitude":
		return "Must be a valid latitude (-90 to 90)"
	case "longitude":
		return "Must be a valid longitude (-180 to 180)"
	case "uuid":
		return "Must be a valid UUID"
	default:
		return fmt.Sprintf("Failed validation: %s", e.Tag())
	}
}
