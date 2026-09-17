package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"ticketflow/pkg/apperr"
	"ticketflow/services/event-service/internal/handler/http/mocks"
	"ticketflow/services/event-service/internal/model"
	"ticketflow/services/event-service/internal/service"
)

const (
	testCaseSuccess_TicketTypeCreate            = "[Success] Body hợp lệ -> gọi Create, trả 201 kèm DTO"
	testCaseError_TicketTypeCreate_InvalidBody  = "[Error] Thiếu field required -> 400 validation_error, không gọi service"
	testCaseError_TicketTypeCreate_ServiceError = "[Error] Service lỗi -> map qua apperr.Respond"
)

type TicketTypeHandlerTestSuite struct {
	suite.Suite
	ticketTypes *mocks.TicketTypeCreator
	handler     *TicketTypeHandler
}

func TestTicketTypeHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(TicketTypeHandlerTestSuite))
}

func (s *TicketTypeHandlerTestSuite) SetupTest()    { s.initMocks() }
func (s *TicketTypeHandlerTestSuite) SetupSubTest() { s.initMocks() }

func (s *TicketTypeHandlerTestSuite) initMocks() {
	s.ticketTypes = mocks.NewTicketTypeCreator(s.T())
	s.handler = NewTicketTypeHandler(s.ticketTypes)
}

func newTicketTypeContext(body []byte, eventID string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/organizer/events/"+eventID+"/ticket-types", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: eventID}}
	return c, w
}

func (s *TicketTypeHandlerTestSuite) TestCreate() {
	s.Run(testCaseSuccess_TicketTypeCreate, func() {
		body, _ := json.Marshal(map[string]any{"name": "VIP", "price": 500, "quota": 10})
		c, w := newTicketTypeContext(body, "e-1")
		s.ticketTypes.On("Create", mock.Anything, "e-1", mock.MatchedBy(func(in service.CreateTicketTypeInput) bool {
			return in.Name == "VIP" && in.Price == 500 && in.Quota == 10
		})).Return(&model.TicketType{ID: "tt-1", EventID: "e-1", Name: "VIP", Price: 500, Quota: 10}, nil)

		s.handler.Create(c)

		s.Equal(http.StatusCreated, w.Code)
	})

	s.Run(testCaseError_TicketTypeCreate_InvalidBody, func() {
		body, _ := json.Marshal(map[string]any{"name": "VIP"}) // missing price/quota
		c, w := newTicketTypeContext(body, "e-1")

		s.handler.Create(c)

		s.Equal(http.StatusBadRequest, w.Code)
	})

	s.Run(testCaseError_TicketTypeCreate_ServiceError, func() {
		body, _ := json.Marshal(map[string]any{"name": "VIP", "price": 500, "quota": 10})
		c, w := newTicketTypeContext(body, "missing-event")
		s.ticketTypes.On("Create", mock.Anything, "missing-event", mock.Anything).Return(nil, apperr.ErrNotFound)

		s.handler.Create(c)

		s.Equal(http.StatusNotFound, w.Code)
	})
}
