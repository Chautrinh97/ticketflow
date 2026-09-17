package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"ticketflow/pkg/apperr"
	"ticketflow/pkg/pagination"
	"ticketflow/services/event-service/internal/handler/http/mocks"
	"ticketflow/services/event-service/internal/model"
	"ticketflow/services/event-service/internal/repository"
	"ticketflow/services/event-service/internal/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

const (
	testCaseSuccess_ListEvents_NoFilters                  = "[Success] Không filter -> trả nguyên danh sách từ service"
	testCaseSuccess_ListEvents_FromFilterParsed           = "[Success] from hợp lệ (RFC3339) -> filter.From được set"
	testCaseSuccess_ListEvents_InvalidFromSilentlyIgnored = "[Success] from không đúng RFC3339 -> bị bỏ qua thay vì lỗi 400"
	testCaseError_ListEvents_ServiceError                 = "[Error] Service lỗi -> map qua apperr.Respond"

	testCaseSuccess_GetBySlug_Found  = "[Success] slug tồn tại -> trả 200 kèm detail DTO"
	testCaseError_GetBySlug_NotFound = "[Error] slug không tồn tại -> map đúng apperr.ErrNotFound"
)

type PublicEventHandlerTestSuite struct {
	suite.Suite
	events  *mocks.EventQueryService
	handler *PublicEventHandler
}

func TestPublicEventHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(PublicEventHandlerTestSuite))
}

func (s *PublicEventHandlerTestSuite) SetupTest()    { s.initMocks() }
func (s *PublicEventHandlerTestSuite) SetupSubTest() { s.initMocks() }

func (s *PublicEventHandlerTestSuite) initMocks() {
	s.events = mocks.NewEventQueryService(s.T())
	s.handler = NewPublicEventHandler(s.events)
}

func newGetContext(rawQuery string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/events?"+rawQuery, nil)
	return c, w
}

func (s *PublicEventHandlerTestSuite) TestListEvents() {
	s.Run(testCaseSuccess_ListEvents_NoFilters, func() {
		c, w := newGetContext("")
		s.events.On("ListPublished", mock.Anything, repository.EventFilter{}, pagination.Params{Page: pagination.DefaultPage, PageSize: pagination.DefaultPageSize}).
			Return([]repository.EventSummaryRow{{ID: "e-1", Title: "T1"}}, int64(1), nil)

		s.handler.ListEvents(c)

		s.Equal(http.StatusOK, w.Code)
	})

	s.Run(testCaseSuccess_ListEvents_FromFilterParsed, func() {
		c, w := newGetContext("category=concert&city=Hanoi&from=2026-01-01T00%3A00%3A00Z")
		wantFrom, _ := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
		s.events.On("ListPublished", mock.Anything, mock.MatchedBy(func(f repository.EventFilter) bool {
			return f.Category == "concert" && f.City == "Hanoi" && f.From != nil && f.From.Equal(wantFrom)
		}), mock.Anything).Return(nil, int64(0), nil)

		s.handler.ListEvents(c)

		s.Equal(http.StatusOK, w.Code)
	})

	s.Run(testCaseSuccess_ListEvents_InvalidFromSilentlyIgnored, func() {
		c, w := newGetContext("from=not-a-date")
		s.events.On("ListPublished", mock.Anything, mock.MatchedBy(func(f repository.EventFilter) bool {
			return f.From == nil
		}), mock.Anything).Return(nil, int64(0), nil)

		s.handler.ListEvents(c)

		s.Equal(http.StatusOK, w.Code)
	})

	s.Run(testCaseError_ListEvents_ServiceError, func() {
		c, w := newGetContext("")
		s.events.On("ListPublished", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, int64(0), apperr.ErrInternal)

		s.handler.ListEvents(c)

		s.Equal(http.StatusInternalServerError, w.Code)
	})
}

func (s *PublicEventHandlerTestSuite) TestGetBySlug() {
	s.Run(testCaseSuccess_GetBySlug_Found, func() {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/events/my-slug", nil)
		c.Params = gin.Params{{Key: "slug", Value: "my-slug"}}
		detail := &service.EventDetail{Event: &model.Event{ID: "e-1", Slug: "my-slug"}}
		s.events.On("GetDetailBySlug", mock.Anything, "my-slug").Return(detail, nil)

		s.handler.GetBySlug(c)

		s.Equal(http.StatusOK, w.Code)
	})

	s.Run(testCaseError_GetBySlug_NotFound, func() {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/events/missing", nil)
		c.Params = gin.Params{{Key: "slug", Value: "missing"}}
		s.events.On("GetDetailBySlug", mock.Anything, "missing").Return(nil, apperr.ErrNotFound)

		s.handler.GetBySlug(c)

		s.Equal(http.StatusNotFound, w.Code)
	})
}
