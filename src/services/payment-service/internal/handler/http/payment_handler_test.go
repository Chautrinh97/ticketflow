package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"ticketflow/pkg/apperr"
	"ticketflow/services/payment-service/internal/handler/http/mocks"
	"ticketflow/services/payment-service/internal/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

const (
	testCaseSuccess_Checkout_EmptyBodyUsesEmptyProvider = "[Success] Body rỗng vẫn checkout được (provider truyền rỗng)"
	testCaseSuccess_Checkout_WithProvider               = "[Success] Body có provider được truyền đúng xuống service"
	testCaseError_Checkout_ServiceAppError              = "[Error] Service trả *apperr.Error được map đúng status/body"
	testCaseError_Checkout_ServiceGenericError          = "[Error] Service trả lỗi thường bị map thành 500 internal_error"

	testCaseSuccess_Webhook_ValidBody           = "[Success] Body hợp lệ gọi HandleWebhook và trả 200 rỗng"
	testCaseError_Webhook_MissingRequiredFields = "[Error] Thiếu order_id/status trả 400 validation_error, không gọi service"
	testCaseError_Webhook_ServiceAppError       = "[Error] Service trả *apperr.Error (vd not_found) được map đúng"
)

type PaymentHandlerTestSuite struct {
	suite.Suite
	useCase *mocks.PaymentUseCase
	handler *PaymentHandler
}

func TestPaymentHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentHandlerTestSuite))
}

func (s *PaymentHandlerTestSuite) SetupTest()    { s.initMocks() }
func (s *PaymentHandlerTestSuite) SetupSubTest() { s.initMocks() }

func (s *PaymentHandlerTestSuite) initMocks() {
	s.useCase = mocks.NewPaymentUseCase(s.T())
	s.handler = NewPaymentHandler(s.useCase)
}

// newTestContext builds a gin.Context with the given method/body, and sets
// :orderId as a route param the way gin's router would after matching
// POST /api/v1/payments/:orderId/checkout.
func newTestContext(method, orderID string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reader *bytes.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	} else {
		reader = bytes.NewReader(nil)
	}
	c.Request = httptest.NewRequest(method, "/api/v1/payments/"+orderID+"/checkout", reader)
	if orderID != "" {
		c.Params = gin.Params{{Key: "orderId", Value: orderID}}
	}
	return c, w
}

func (s *PaymentHandlerTestSuite) TestCheckout() {
	s.Run(testCaseSuccess_Checkout_EmptyBodyUsesEmptyProvider, func() {
		const orderID = "order-1"
		s.useCase.On("Checkout", mock.Anything, orderID, "").
			Return(&service.CheckoutResult{PaymentID: "pay-1", CheckoutURL: "/orders/order-1/confirmation"}, nil)

		c, w := newTestContext(http.MethodPost, orderID, nil)
		s.handler.Checkout(c)

		s.Require().Equal(http.StatusOK, w.Code)
		var got checkoutResponseDTO
		s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &got))
		s.Assert().Equal("pay-1", got.PaymentID)
		s.Assert().Equal("/orders/order-1/confirmation", got.CheckoutURL)
	})

	s.Run(testCaseSuccess_Checkout_WithProvider, func() {
		const orderID = "order-2"
		s.useCase.On("Checkout", mock.Anything, orderID, "vnpay").
			Return(&service.CheckoutResult{PaymentID: "pay-2", CheckoutURL: "/orders/order-2/confirmation"}, nil)

		body, err := json.Marshal(checkoutRequest{Provider: "vnpay"})
		s.Require().NoError(err)
		c, w := newTestContext(http.MethodPost, orderID, body)
		s.handler.Checkout(c)

		s.Require().Equal(http.StatusOK, w.Code)
	})

	s.Run(testCaseError_Checkout_ServiceAppError, func() {
		const orderID = "order-3"
		s.useCase.On("Checkout", mock.Anything, orderID, "").
			Return(nil, apperr.WithMessage(apperr.ErrConflict, "Đơn hàng không ở trạng thái chờ thanh toán"))

		c, w := newTestContext(http.MethodPost, orderID, nil)
		s.handler.Checkout(c)

		s.Require().Equal(http.StatusConflict, w.Code)
		s.Assert().JSONEq(`{"code":"conflict","message":"Đơn hàng không ở trạng thái chờ thanh toán"}`, w.Body.String())
	})

	s.Run(testCaseError_Checkout_ServiceGenericError, func() {
		const orderID = "order-4"
		s.useCase.On("Checkout", mock.Anything, orderID, "").
			Return(nil, errors.New("some raw internal detail"))

		c, w := newTestContext(http.MethodPost, orderID, nil)
		s.handler.Checkout(c)

		s.Require().Equal(http.StatusInternalServerError, w.Code)
		s.Assert().NotContains(w.Body.String(), "raw internal detail")
	})
}

func (s *PaymentHandlerTestSuite) TestWebhook() {
	s.Run(testCaseSuccess_Webhook_ValidBody, func() {
		s.useCase.On("HandleWebhook", mock.Anything, "order-1", "success").Return(nil)

		body, err := json.Marshal(webhookRequest{OrderID: "order-1", Status: "success"})
		s.Require().NoError(err)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", bytes.NewReader(body))

		s.handler.Webhook(c)

		s.Require().Equal(http.StatusOK, w.Code)
		s.Assert().Empty(w.Body.String())
	})

	s.Run(testCaseError_Webhook_MissingRequiredFields, func() {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", bytes.NewReader([]byte(`{}`)))

		s.handler.Webhook(c)

		s.Require().Equal(http.StatusBadRequest, w.Code)
		s.Assert().JSONEq(`{"code":"validation_error","message":"order_id và status là bắt buộc"}`, w.Body.String())
		s.useCase.AssertNotCalled(s.T(), "HandleWebhook", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run(testCaseError_Webhook_ServiceAppError, func() {
		s.useCase.On("HandleWebhook", mock.Anything, "order-2", "success").Return(apperr.ErrNotFound)

		body, err := json.Marshal(webhookRequest{OrderID: "order-2", Status: "success"})
		s.Require().NoError(err)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", bytes.NewReader(body))

		s.handler.Webhook(c)

		s.Require().Equal(http.StatusNotFound, w.Code)
	})
}
