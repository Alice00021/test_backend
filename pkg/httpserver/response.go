package httpserver

import (
	"errors"
	"fmt"
	"net/http"
)

// HttpError - represents an HTTP error response.
type HttpError struct {
	Status  int         `json:"-"`
	Message string      `json:"error"`
	Details interface{} `json:"details,omitempty"`
}

// Error - .
func (e HttpError) Error() string {
	return fmt.Sprintf("status: %d - errors: %s - details: %v", e.Status, e.Message, e.Details)
}

// NewHttpError - .
func NewHttpError(status int, message string, details interface{}) HttpError {
	httpError := HttpError{
		Status:  status,
		Message: message,
	}

	err, ok := details.(error)
	if ok {
		sourceError := func(err error) error {
			var errs []error
			for err != nil {
				errs = append(errs, err)
				err = errors.Unwrap(err)
			}
			if len(errs) == 0 {
				return err
			}

			return errs[len(errs)-1]
		}

		httpError.Details = sourceError(err).Error()
	} else {
		httpError.Details = details
	}

	return httpError
}

// NewBadRequestError - .
func NewBadRequestError(details interface{}) HttpError {
	return NewHttpError(http.StatusBadRequest, "Bad request", details)
}

// NewBadQueryParamsError - .
func NewBadQueryParamsError(details interface{}) HttpError {
	return NewHttpError(http.StatusBadRequest, "Invalid query params", details)
}

// NewBadRequestBodyError - .
func NewBadRequestBodyError(details interface{}) HttpError {
	return NewHttpError(http.StatusBadRequest, "Invalid request body", details)
}

// NewBadPathParamsError - .
func NewBadPathParamsError(details interface{}) HttpError {
	return NewHttpError(http.StatusBadRequest, "Invalid path params", details)
}

// NewUnauthorizedError - .
func NewUnauthorizedError(details interface{}) HttpError {
	return NewHttpError(http.StatusUnauthorized, "Unauthorized", details)
}

// NewForbiddenError - .
func NewForbiddenError(details interface{}) HttpError {
	return NewHttpError(http.StatusForbidden, "Forbidden", details)
}

// NewInternalServerError - .
func NewInternalServerError(details interface{}) HttpError {
	return NewHttpError(http.StatusInternalServerError, "Internal Server Error", details)
}

// NewNotFoundError - .
func NewNotFoundError(details interface{}) HttpError {
	return NewHttpError(http.StatusNotFound, "Not Found", details)
}

// NewConflictError - .
func NewConflictError(details interface{}) HttpError {
	return NewHttpError(http.StatusConflict, "Conflict", details)
}
