package errorsx

import (
	"encoding/json"
	"errors"
	"net/http"
)

const (
	CodeInvalidArgument = "INVALID_ARGUMENT"
	CodeNotFound        = "NOT_FOUND"
	CodeConflict        = "CONFLICT"
	CodeWaitTimeout     = "WAIT_TIMEOUT"
	CodeCanceled        = "CANCELED"
	CodeDemoDisabled    = "DEMO_DISABLED"
	CodeNotReady        = "NOT_READY"
	CodeInternal        = "INTERNAL"
	CodePoolClosed      = "POOL_CLOSED"
)

type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
	HTTP    int            `json:"-"`
	err     error          `json:"-"`
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Code + ": " + e.Message
}

func (e *Error) Unwrap() error { return e.err }

func New(code, msg string, httpStatus int) *Error {
	return &Error{Code: code, Message: msg, HTTP: httpStatus, Details: map[string]any{}}
}

func WithDetail(err *Error, key string, value any) *Error {
	if err.Details == nil {
		err.Details = map[string]any{}
	}
	err.Details[key] = value
	return err
}

func Wrap(code, msg string, httpStatus int, err error) *Error {
	e := New(code, msg, httpStatus)
	e.err = err
	return e
}

func Invalid(field, msg string) *Error {
	return WithDetail(New(CodeInvalidArgument, msg, http.StatusBadRequest), "field", field)
}

func HTTPStatus(err error) int {
	var e *Error
	if errors.As(err, &e) && e.HTTP != 0 {
		return e.HTTP
	}
	return http.StatusInternalServerError
}

func CodeOf(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return CodeInternal
}

type envelope struct {
	Error *Error `json:"error"`
}

func Write(w http.ResponseWriter, err error) {
	e := &Error{Code: CodeInternal, Message: "internal error", HTTP: http.StatusInternalServerError}
	var typed *Error
	if errors.As(err, &typed) {
		e = typed
	}
	if e.HTTP == 0 {
		e.HTTP = http.StatusInternalServerError
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.HTTP)
	_ = json.NewEncoder(w).Encode(envelope{Error: e})
}
