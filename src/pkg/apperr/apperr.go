// Package apperr defines the canonical error shape shared by every service,
// mirroring the {code, message} Error schema used identically across all
// api-docs/openapi/*.yaml files, plus HTTP/gRPC status mapping helpers.
package apperr

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Error struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *Error) Error() string { return e.Message }

func New(httpStatus int, code, message string) *Error {
	return &Error{Code: code, Message: message, HTTPStatus: httpStatus}
}

// WithMessage returns a copy of base with a more specific message, keeping
// its code/HTTPStatus — use for template errors like ErrValidation.
func WithMessage(base *Error, message string) *Error {
	return &Error{Code: base.Code, Message: message, HTTPStatus: base.HTTPStatus}
}

var (
	ErrUnauthorized    = New(http.StatusUnauthorized, "unauthorized", "Yêu cầu đăng nhập")
	ErrForbidden       = New(http.StatusForbidden, "forbidden", "Không có quyền thực hiện hành động này")
	ErrNotFound        = New(http.StatusNotFound, "not_found", "Không tìm thấy tài nguyên")
	ErrConflict        = New(http.StatusConflict, "conflict", "Xung đột dữ liệu")
	ErrTooManyRequests = New(http.StatusTooManyRequests, "rate_limited", "Vượt giới hạn tần suất")
	ErrValidation      = New(http.StatusBadRequest, "validation_error", "Dữ liệu không hợp lệ")
	ErrInternal        = New(http.StatusInternalServerError, "internal_error", "Lỗi hệ thống")
)

// JSON writes {code,message} with the error's own HTTP status onto c.
func JSON(c *gin.Context, err *Error) {
	c.JSON(err.HTTPStatus, err)
}

// Respond maps any error to a response: *Error is written as-is, anything
// else becomes ErrInternal (never leak raw internal error strings to clients).
func Respond(c *gin.Context, err error) {
	var appErr *Error
	if errors.As(err, &appErr) {
		JSON(c, appErr)
		return
	}
	JSON(c, ErrInternal)
}

// grpcCodeFor maps an *Error's HTTP status to the closest gRPC status code,
// for services that also expose the same error over an internal RPC.
func grpcCodeFor(httpStatus int) codes.Code {
	switch httpStatus {
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.AlreadyExists
	case http.StatusTooManyRequests:
		return codes.ResourceExhausted
	case http.StatusBadRequest:
		return codes.InvalidArgument
	default:
		return codes.Internal
	}
}

// ToGRPCStatus converts an *Error into a gRPC status error, encoding Code
// as the status message prefix so callers can recover it via FromGRPCError.
func ToGRPCStatus(err *Error) error {
	return status.Error(grpcCodeFor(err.HTTPStatus), err.Code+": "+err.Message)
}

// AsGRPCStatus converts any error returned by a service-layer call into a
// gRPC status error for a gRPC server handler to return — *Error is mapped
// precisely via ToGRPCStatus, anything else becomes an internal error.
func AsGRPCStatus(err error) error {
	var appErr *Error
	if errors.As(err, &appErr) {
		return ToGRPCStatus(appErr)
	}
	return ToGRPCStatus(ErrInternal)
}

// FromGRPCError converts a gRPC error (from an internal service call) back
// into an *Error suitable for an HTTP handler to return to its own client.
func FromGRPCError(err error) *Error {
	st, ok := status.FromError(err)
	if !ok {
		return ErrInternal
	}
	switch st.Code() {
	case codes.Unauthenticated:
		return WithMessage(ErrUnauthorized, st.Message())
	case codes.PermissionDenied:
		return WithMessage(ErrForbidden, st.Message())
	case codes.NotFound:
		return WithMessage(ErrNotFound, st.Message())
	case codes.AlreadyExists:
		return WithMessage(ErrConflict, st.Message())
	case codes.ResourceExhausted:
		return WithMessage(ErrTooManyRequests, st.Message())
	case codes.InvalidArgument:
		return WithMessage(ErrValidation, st.Message())
	default:
		return ErrInternal
	}
}
