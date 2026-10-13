package httpx

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func OK(w http.ResponseWriter, v any) {
	respond(w, http.StatusOK, v)
}

func Created(w http.ResponseWriter, v any) {
	respond(w, http.StatusCreated, v)
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Error(w http.ResponseWriter, err ErrorDetail) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatus())
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: err})
}

func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

var (
	ErrInvalidJSON = ErrorDetail{Code: "INVALID_JSON", Message: "invalid JSON body"}
	ErrInvalidID   = ErrorDetail{Code: "INVALID_ID", Message: "invalid task ID format"}
	ErrNotFound    = ErrorDetail{Code: "NOT_FOUND", Message: "resource not found"}
	ErrInternal    = ErrorDetail{Code: "INTERNAL_ERROR", Message: "internal server error"}
)

func ErrValidation(field, message string) ErrorDetail {
	return ErrorDetail{Code: "VALIDATION_ERROR", Message: message, Field: field}
}

func ErrInvalidQueryParam(param, message string) ErrorDetail {
	return ErrorDetail{Code: "INVALID_QUERY_PARAM", Message: param + ": " + message, Field: param}
}

func (e ErrorDetail) HTTPStatus() int {
	switch e.Code {
	case "INVALID_JSON", "INVALID_ID", "VALIDATION_ERROR", "INVALID_QUERY_PARAM":
		return http.StatusBadRequest
	case "NOT_FOUND":
		return http.StatusNotFound
	case "REQUEST_TIMEOUT":
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}
