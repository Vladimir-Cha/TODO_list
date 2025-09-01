package errors

import (
	"fmt"
	"net/http"
)

type Error struct {
	Code    int    // http статус
	Message string // ообщение об ошибке
	Details any    // дополнительные данные (мапы для json ответов)
	Err     error  // исходная ошибка
}

var (
	ErrNotFound = &Error{
		Code:    http.StatusNotFound,
		Message: "Not found",
	}

	ErrBadRequest = &Error{
		Code:    http.StatusBadRequest,
		Message: "Status bad request",
	}

	ErrDatabase = &Error{
		Code:    http.StatusInternalServerError,
		Message: "Internal server error",
	}

	ErrValidation = &Error{
		Code:    http.StatusUnprocessableEntity,
		Message: "Validation failed",
	}
)

// ошибка с деталями
func (e *Error) WithDetails(details any) *Error {
	return &Error{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
	}
}

// ошибка с кодом, сообщением и исходной ошибкой
func (e *Error) WithError(err error) *Error {
	return &Error{
		Code:    e.Code,
		Message: e.Message,
	}
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// ошибка для json
func (e *Error) WithMap() map[string]any {
	m := map[string]any{
		"error": e.Message,
	}
	if e.Details != nil {
		m["details"] = e.Details
	}
	if e.Err != nil {
		m["internal"] = e.Err.Error()
	}
	return m
}
