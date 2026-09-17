package http

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"ticketflow/pkg/apperr"
	"ticketflow/pkg/httpauth"
	"ticketflow/pkg/pagination"
	"ticketflow/services/booking-service/internal/handler/http/mocks"
	"ticketflow/services/booking-service/internal/model"
)

const (
	testCaseSuccess_Create_ReturnsCreatedOrder = "[Success] Trả 201 kèm order khi tạo thành công"
	testCaseError_Create_InvalidBody           = "[Error] Body thiếu items -> 400, không gọi service"
	testCaseError_Create_ServiceError          = "[Error] Lỗi từ service map đúng qua apperr.Respond"

	testCaseSuccess_GetByID_ReturnsOrder = "[Success] Trả 200 kèm order"
	testCaseError_GetByID_NotFound       = "[Error] Order không tồn tại -> 404"

	testCaseSuccess_ListMine_ReturnsEnvelope = "[Success] Trả {items,total} đúng phân trang"
	testCaseError_ListMine_ServiceError      = "[Error] Lỗi từ service map đúng qua apperr.Respond"

	testCaseSuccess_OwnerLookup_ReturnsOwnerID = "[Success] Trả về ownerID từ service.GetOwnerID"
	testCaseError_OwnerLookup_ServiceError     = "[Error] Lan truyền lỗi khi order không tồn tại"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type BookingHandlerTestSuite struct {
	suite.Suite
}

func TestBookingHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(BookingHandlerTestSuite))
}

func newTestContext(method, path string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reader *bytes.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return c, w
}

func (s *BookingHandlerTestSuite) TestCreate() {
	s.Run(testCaseSuccess_Create_ReturnsCreatedOrder, func() {
		bookings := mocks.NewBookings(s.T())
		h := NewBookingHandler(bookings)
		c, w := newTestContext(http.MethodPost, "/api/v1/bookings",
			[]byte(`{"items":[{"ticket_type_id":"tt-1","quantity":2}]}`))
		c.Set(httpauth.ContextUserID, "user-1")

		wantOrder := &model.Order{ID: "order-1", UserID: "user-1", Status: model.OrderStatusPending}
		bookings.On("CreateOrder", mock.Anything, "user-1", []model.BookingItem{{TicketTypeID: "tt-1", Quantity: 2}}).
			Return(wantOrder, nil)

		h.Create(c)

		s.Require().Equal(http.StatusCreated, w.Code)
		s.Assert().Contains(w.Body.String(), `"id":"order-1"`)
	})

	s.Run(testCaseError_Create_InvalidBody, func() {
		bookings := mocks.NewBookings(s.T())
		h := NewBookingHandler(bookings)
		c, w := newTestContext(http.MethodPost, "/api/v1/bookings", []byte(`{}`))
		c.Set(httpauth.ContextUserID, "user-1")

		h.Create(c)

		s.Require().Equal(http.StatusBadRequest, w.Code)
		bookings.AssertNotCalled(s.T(), "CreateOrder", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run(testCaseError_Create_ServiceError, func() {
		bookings := mocks.NewBookings(s.T())
		h := NewBookingHandler(bookings)
		c, w := newTestContext(http.MethodPost, "/api/v1/bookings",
			[]byte(`{"items":[{"ticket_type_id":"tt-1","quantity":1}]}`))
		c.Set(httpauth.ContextUserID, "user-1")

		bookings.On("CreateOrder", mock.Anything, "user-1", mock.Anything).
			Return(nil, apperr.WithMessage(apperr.ErrConflict, "Không đủ tồn kho"))

		h.Create(c)

		s.Require().Equal(http.StatusConflict, w.Code)
		s.Assert().Contains(w.Body.String(), "Không đủ tồn kho")
	})
}

func (s *BookingHandlerTestSuite) TestGetByID() {
	s.Run(testCaseSuccess_GetByID_ReturnsOrder, func() {
		bookings := mocks.NewBookings(s.T())
		h := NewBookingHandler(bookings)
		c, w := newTestContext(http.MethodGet, "/api/v1/bookings/order-1", nil)
		c.Params = gin.Params{{Key: "id", Value: "order-1"}}

		bookings.On("GetOrderByID", mock.Anything, "order-1").
			Return(&model.Order{ID: "order-1", UserID: "user-1"}, nil)

		h.GetByID(c)

		s.Require().Equal(http.StatusOK, w.Code)
		s.Assert().Contains(w.Body.String(), `"id":"order-1"`)
	})

	s.Run(testCaseError_GetByID_NotFound, func() {
		bookings := mocks.NewBookings(s.T())
		h := NewBookingHandler(bookings)
		c, w := newTestContext(http.MethodGet, "/api/v1/bookings/missing", nil)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}

		bookings.On("GetOrderByID", mock.Anything, "missing").Return(nil, apperr.ErrNotFound)

		h.GetByID(c)

		s.Require().Equal(http.StatusNotFound, w.Code)
	})
}

func (s *BookingHandlerTestSuite) TestListMine() {
	s.Run(testCaseSuccess_ListMine_ReturnsEnvelope, func() {
		bookings := mocks.NewBookings(s.T())
		h := NewBookingHandler(bookings)
		c, w := newTestContext(http.MethodGet, "/api/v1/users/me/bookings?page=2&page_size=5&status=paid", nil)
		c.Set(httpauth.ContextUserID, "user-1")

		orders := []*model.Order{{ID: "o1", Status: model.OrderStatusPaid}}
		bookings.On("ListMyOrders", mock.Anything, "user-1", "paid", pagination.Params{Page: 2, PageSize: 5}).
			Return(orders, int64(1), nil)

		h.ListMine(c)

		s.Require().Equal(http.StatusOK, w.Code)
		s.Assert().Contains(w.Body.String(), `"total":1`)
		s.Assert().Contains(w.Body.String(), `"id":"o1"`)
	})

	s.Run(testCaseError_ListMine_ServiceError, func() {
		bookings := mocks.NewBookings(s.T())
		h := NewBookingHandler(bookings)
		c, w := newTestContext(http.MethodGet, "/api/v1/users/me/bookings", nil)
		c.Set(httpauth.ContextUserID, "user-1")

		bookings.On("ListMyOrders", mock.Anything, "user-1", "", mock.Anything).
			Return(nil, int64(0), errors.New("db down"))

		h.ListMine(c)

		s.Require().Equal(http.StatusInternalServerError, w.Code)
		s.Assert().NotContains(w.Body.String(), "db down")
	})
}

func (s *BookingHandlerTestSuite) TestOwnerLookup() {
	s.Run(testCaseSuccess_OwnerLookup_ReturnsOwnerID, func() {
		bookings := mocks.NewBookings(s.T())
		h := NewBookingHandler(bookings)
		c, _ := newTestContext(http.MethodGet, "/api/v1/bookings/order-1", nil)
		c.Params = gin.Params{{Key: "id", Value: "order-1"}}

		bookings.On("GetOwnerID", mock.Anything, "order-1").Return("user-1", nil)

		owner, err := h.OwnerLookup(c)

		s.Require().NoError(err)
		s.Assert().Equal("user-1", owner)
	})

	s.Run(testCaseError_OwnerLookup_ServiceError, func() {
		bookings := mocks.NewBookings(s.T())
		h := NewBookingHandler(bookings)
		c, _ := newTestContext(http.MethodGet, "/api/v1/bookings/missing", nil)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}

		bookings.On("GetOwnerID", mock.Anything, "missing").Return("", apperr.ErrNotFound)

		owner, err := h.OwnerLookup(c)

		s.Require().Error(err)
		s.Assert().Empty(owner)
	})
}
