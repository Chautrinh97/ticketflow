package apperr

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestNew(t *testing.T) {
	err := New(http.StatusTeapot, "teapot", "I'm a teapot")
	assert.Equal(t, "teapot", err.Code)
	assert.Equal(t, "I'm a teapot", err.Message)
	assert.Equal(t, http.StatusTeapot, err.HTTPStatus)
}

func TestError_Error(t *testing.T) {
	err := New(http.StatusBadRequest, "bad", "something bad")
	assert.Equal(t, "something bad", err.Error())
}

func TestWithMessage(t *testing.T) {
	specific := WithMessage(ErrValidation, "email không hợp lệ")
	assert.Equal(t, ErrValidation.Code, specific.Code)
	assert.Equal(t, ErrValidation.HTTPStatus, specific.HTTPStatus)
	assert.Equal(t, "email không hợp lệ", specific.Message)
	// base var must stay untouched — WithMessage returns a copy.
	assert.Equal(t, "Dữ liệu không hợp lệ", ErrValidation.Message)
	assert.NotSame(t, ErrValidation, specific)
}

func TestPredefinedErrors(t *testing.T) {
	cases := []struct {
		name       string
		err        *Error
		wantCode   string
		wantStatus int
	}{
		{"Unauthorized", ErrUnauthorized, "unauthorized", http.StatusUnauthorized},
		{"Forbidden", ErrForbidden, "forbidden", http.StatusForbidden},
		{"NotFound", ErrNotFound, "not_found", http.StatusNotFound},
		{"Conflict", ErrConflict, "conflict", http.StatusConflict},
		{"TooManyRequests", ErrTooManyRequests, "rate_limited", http.StatusTooManyRequests},
		{"Validation", ErrValidation, "validation_error", http.StatusBadRequest},
		{"Internal", ErrInternal, "internal_error", http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantCode, tc.err.Code)
			assert.Equal(t, tc.wantStatus, tc.err.HTTPStatus)
		})
	}
}

func TestJSON(t *testing.T) {
	c, w := newTestGinContext()
	JSON(c, ErrNotFound)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"code":"not_found","message":"Không tìm thấy tài nguyên"}`, w.Body.String())
}

func TestRespond(t *testing.T) {
	t.Run("[Success] *Error passed through as-is", func(t *testing.T) {
		c, w := newTestGinContext()
		Respond(c, ErrConflict)
		assert.Equal(t, http.StatusConflict, w.Code)
		assert.JSONEq(t, `{"code":"conflict","message":"Xung đột dữ liệu"}`, w.Body.String())
	})

	t.Run("[Success] wrapped *Error unwrapped via errors.As", func(t *testing.T) {
		c, w := newTestGinContext()
		wrapped := fmt.Errorf("repository: %w", ErrConflict)
		Respond(c, wrapped)
		assert.Equal(t, http.StatusConflict, w.Code)
		assert.JSONEq(t, `{"code":"conflict","message":"Xung đột dữ liệu"}`, w.Body.String())
	})

	t.Run("[Error] plain error maps to ErrInternal without leaking its message", func(t *testing.T) {
		c, w := newTestGinContext()
		Respond(c, errors.New("some raw internal detail"))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"code":"internal_error","message":"Lỗi hệ thống"}`, w.Body.String())
		assert.NotContains(t, w.Body.String(), "raw internal detail")
	})
}

func TestToGRPCStatus_FromGRPCError_RoundTrip(t *testing.T) {
	// FromGRPCError reconstructs the message from the gRPC status message,
	// which ToGRPCStatus encoded as "code: message" — so the round-tripped
	// message is NOT byte-identical to the original base.Message, it carries
	// that prefix. Assert the real behavior, not a naive clean round trip.
	cases := []*Error{ErrUnauthorized, ErrForbidden, ErrNotFound, ErrConflict, ErrTooManyRequests, ErrValidation}
	for _, base := range cases {
		t.Run(base.Code, func(t *testing.T) {
			grpcErr := ToGRPCStatus(base)
			back := FromGRPCError(grpcErr)
			assert.Equal(t, base.Code, back.Code)
			assert.Equal(t, base.HTTPStatus, back.HTTPStatus)
			assert.Equal(t, base.Code+": "+base.Message, back.Message)
		})
	}
}

func TestFromGRPCError_UnmappedCodeFallsBackToInternal(t *testing.T) {
	grpcErr := status.Error(codes.Unavailable, "dial tcp: connection refused")
	back := FromGRPCError(grpcErr)
	assert.Equal(t, ErrInternal.Code, back.Code)
	assert.Equal(t, ErrInternal.HTTPStatus, back.HTTPStatus)
}

func TestFromGRPCError_NonStatusErrorFallsBackToInternal(t *testing.T) {
	back := FromGRPCError(errors.New("not a grpc status error"))
	assert.Equal(t, ErrInternal.Code, back.Code)
	assert.Equal(t, ErrInternal.HTTPStatus, back.HTTPStatus)
}

func TestAsGRPCStatus(t *testing.T) {
	t.Run("[Success] *Error maps to its own grpc status code", func(t *testing.T) {
		got := AsGRPCStatus(ErrNotFound)
		st, ok := status.FromError(got)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("[Error] generic error maps to Internal", func(t *testing.T) {
		got := AsGRPCStatus(errors.New("boom"))
		st, ok := status.FromError(got)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
	})
}
