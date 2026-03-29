package errors

import (
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorType int

const (
	ErrorTypeUnknown ErrorType = iota
	ErrorTypeValidation
	ErrorTypeNotFound
	ErrorTypeAlreadyExists
	ErrorTypeUnauthorized
	ErrorTypeForbidden
	ErrorTypeInternal
	ErrorTypeExternal
	ErrorTypeTimeout
	ErrorTypeRateLimit
)

type AppError struct {
	Type    ErrorType
	Code    string
	Message string
	Details map[string]interface{}
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func (e *AppError) WithDetails(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

func (e *AppError) ToGRPCStatus() *status.Status {
	switch e.Type {
	case ErrorTypeValidation:
		return status.New(codes.InvalidArgument, e.Message)
	case ErrorTypeNotFound:
		return status.New(codes.NotFound, e.Message)
	case ErrorTypeAlreadyExists:
		return status.New(codes.AlreadyExists, e.Message)
	case ErrorTypeUnauthorized:
		return status.New(codes.Unauthenticated, e.Message)
	case ErrorTypeForbidden:
		return status.New(codes.PermissionDenied, e.Message)
	case ErrorTypeTimeout:
		return status.New(codes.DeadlineExceeded, e.Message)
	case ErrorTypeRateLimit:
		return status.New(codes.ResourceExhausted, e.Message)
	case ErrorTypeExternal:
		return status.New(codes.Unavailable, e.Message)
	default:
		return status.New(codes.Internal, e.Message)
	}
}

func (e *AppError) ToHTTPStatus() int {
	switch e.Type {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeAlreadyExists:
		return http.StatusConflict
	case ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case ErrorTypeForbidden:
		return http.StatusForbidden
	case ErrorTypeTimeout:
		return http.StatusRequestTimeout
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	case ErrorTypeExternal:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

func NewValidationError(message string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Code:    "VALIDATION_ERROR",
		Message: message,
		Cause:   cause,
	}
}

func NewNotFoundError(resource string, id string) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s with id '%s' not found", resource, id),
	}
}

func NewAlreadyExistsError(resource string, id string) *AppError {
	return &AppError{
		Type:    ErrorTypeAlreadyExists,
		Code:    "ALREADY_EXISTS",
		Message: fmt.Sprintf("%s with id '%s' already exists", resource, id),
	}
}

func NewInternalError(message string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeInternal,
		Code:    "INTERNAL_ERROR",
		Message: message,
		Cause:   cause,
	}
}

func NewExternalError(service string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeExternal,
		Code:    "EXTERNAL_ERROR",
		Message: fmt.Sprintf("external service '%s' error", service),
		Cause:   cause,
	}
}

func NewTimeoutError(operation string) *AppError {
	return &AppError{
		Type:    ErrorTypeTimeout,
		Code:    "TIMEOUT_ERROR",
		Message: fmt.Sprintf("operation '%s' timed out", operation),
	}
}

func WrapError(err error, message string) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Type:    appErr.Type,
			Code:    appErr.Code,
			Message: message,
			Cause:   appErr,
		}
	}

	return &AppError{
		Type:    ErrorTypeInternal,
		Code:    "WRAPPED_ERROR",
		Message: message,
		Cause:   err,
	}
}

func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

func AsAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}
