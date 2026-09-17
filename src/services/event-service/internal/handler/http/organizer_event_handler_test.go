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
	"ticketflow/pkg/httpauth"
	"ticketflow/pkg/pagination"
	"ticketflow/services/event-service/internal/handler/http/mocks"
	"ticketflow/services/event-service/internal/model"
	"ticketflow/services/event-service/internal/repository"
	"ticketflow/services/event-service/internal/service"
)

const (
	testCaseSuccess_OrganizerList            = "[Success] List trả nguyên kết quả từ service kèm userID từ context"
	testCaseError_OrganizerList_ServiceError = "[Error] Service lỗi -> map qua apperr.Respond"

	testCaseSuccess_OrganizerCreate            = "[Success] Body hợp lệ -> gọi Create với userID từ context, trả 201"
	testCaseError_OrganizerCreate_InvalidBody  = "[Error] Body thiếu field required -> 400 validation_error, không gọi service"
	testCaseError_OrganizerCreate_ServiceError = "[Error] Service lỗi -> map qua apperr.Respond"

	testCaseSuccess_OrganizerGetByID        = "[Success] GetByID trả 200 kèm detail"
	testCaseError_OrganizerGetByID_NotFound = "[Error] Service trả not_found -> map đúng status"

	testCaseSuccess_OrganizerUpdate           = "[Success] Body hợp lệ -> gọi Update, trả 200"
	testCaseError_OrganizerUpdate_InvalidBody = "[Error] Body JSON không hợp lệ -> 400 validation_error"

	testCaseSuccess_OrganizerDelete            = "[Success] Cancel thành công -> 204 no content"
	testCaseError_OrganizerDelete_ServiceError = "[Error] Service lỗi -> map qua apperr.Respond"

	testCaseSuccess_OrganizerPublish              = "[Success] Publish rồi GetDetailByID -> 200 kèm detail mới"
	testCaseError_OrganizerPublish_PublishFails   = "[Error] Publish lỗi -> map lỗi, không gọi GetDetailByID"
	testCaseError_OrganizerPublish_GetDetailFails = "[Error] Publish thành công nhưng GetDetailByID lỗi -> map lỗi"

	testCaseSuccess_OrganizerOwnerLookup = "[Success] OwnerLookup chuyển tiếp đúng kết quả từ GetOwnerID"
)

type OrganizerEventHandlerTestSuite struct {
	suite.Suite
	events  *mocks.OrganizerEventService
	handler *OrganizerEventHandler
}

func TestOrganizerEventHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizerEventHandlerTestSuite))
}

func (s *OrganizerEventHandlerTestSuite) SetupTest()    { s.initMocks() }
func (s *OrganizerEventHandlerTestSuite) SetupSubTest() { s.initMocks() }

func (s *OrganizerEventHandlerTestSuite) initMocks() {
	s.events = mocks.NewOrganizerEventService(s.T())
	s.handler = NewOrganizerEventHandler(s.events)
}

func newOrganizerContext(method, path string, body []byte, params gin.Params) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reader *bytes.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	} else {
		reader = bytes.NewReader(nil)
	}
	c.Request = httptest.NewRequest(method, path, reader)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = params
	c.Set(httpauth.ContextUserID, "user-1")
	return c, w
}

func (s *OrganizerEventHandlerTestSuite) TestList() {
	s.Run(testCaseSuccess_OrganizerList, func() {
		c, w := newOrganizerContext(http.MethodGet, "/organizer/events", nil, nil)
		s.events.On("ListByOrganizer", mock.Anything, "user-1", "", pagination.Params{Page: pagination.DefaultPage, PageSize: pagination.DefaultPageSize}).
			Return([]repository.EventSummaryRow{{ID: "e-1"}}, int64(1), nil)

		s.handler.List(c)

		s.Equal(http.StatusOK, w.Code)
	})

	s.Run(testCaseError_OrganizerList_ServiceError, func() {
		c, w := newOrganizerContext(http.MethodGet, "/organizer/events", nil, nil)
		s.events.On("ListByOrganizer", mock.Anything, "user-1", "", mock.Anything).
			Return(nil, int64(0), apperr.ErrInternal)

		s.handler.List(c)

		s.Equal(http.StatusInternalServerError, w.Code)
	})
}

func (s *OrganizerEventHandlerTestSuite) TestCreate() {
	s.Run(testCaseSuccess_OrganizerCreate, func() {
		body, _ := json.Marshal(map[string]any{
			"title": "New Event", "category": "concert", "start_time": "2026-01-01T00:00:00Z",
		})
		c, w := newOrganizerContext(http.MethodPost, "/organizer/events", body, nil)
		detail := &service.EventDetail{Event: &model.Event{ID: "e-1", OrganizerID: "user-1"}}
		s.events.On("Create", mock.Anything, "user-1", mock.MatchedBy(func(in service.CreateEventInput) bool {
			return in.Title == "New Event" && in.Category == "concert"
		})).Return(detail, nil)

		s.handler.Create(c)

		s.Equal(http.StatusCreated, w.Code)
	})

	s.Run(testCaseError_OrganizerCreate_InvalidBody, func() {
		body, _ := json.Marshal(map[string]any{"title": "Missing category and start_time"})
		c, w := newOrganizerContext(http.MethodPost, "/organizer/events", body, nil)

		s.handler.Create(c)

		s.Equal(http.StatusBadRequest, w.Code)
	})

	s.Run(testCaseError_OrganizerCreate_ServiceError, func() {
		body, _ := json.Marshal(map[string]any{
			"title": "New Event", "category": "concert", "start_time": "2026-01-01T00:00:00Z",
		})
		c, w := newOrganizerContext(http.MethodPost, "/organizer/events", body, nil)
		s.events.On("Create", mock.Anything, "user-1", mock.Anything).Return(nil, apperr.ErrInternal)

		s.handler.Create(c)

		s.Equal(http.StatusInternalServerError, w.Code)
	})
}

func (s *OrganizerEventHandlerTestSuite) TestGetByID() {
	s.Run(testCaseSuccess_OrganizerGetByID, func() {
		c, w := newOrganizerContext(http.MethodGet, "/organizer/events/e-1", nil, gin.Params{{Key: "id", Value: "e-1"}})
		detail := &service.EventDetail{Event: &model.Event{ID: "e-1"}}
		s.events.On("GetDetailByID", mock.Anything, "e-1").Return(detail, nil)

		s.handler.GetByID(c)

		s.Equal(http.StatusOK, w.Code)
	})

	s.Run(testCaseError_OrganizerGetByID_NotFound, func() {
		c, w := newOrganizerContext(http.MethodGet, "/organizer/events/missing", nil, gin.Params{{Key: "id", Value: "missing"}})
		s.events.On("GetDetailByID", mock.Anything, "missing").Return(nil, apperr.ErrNotFound)

		s.handler.GetByID(c)

		s.Equal(http.StatusNotFound, w.Code)
	})
}

func (s *OrganizerEventHandlerTestSuite) TestUpdate() {
	s.Run(testCaseSuccess_OrganizerUpdate, func() {
		body, _ := json.Marshal(map[string]any{"title": "Updated"})
		c, w := newOrganizerContext(http.MethodPut, "/organizer/events/e-1", body, gin.Params{{Key: "id", Value: "e-1"}})
		detail := &service.EventDetail{Event: &model.Event{ID: "e-1"}}
		s.events.On("Update", mock.Anything, "e-1", mock.Anything).Return(detail, nil)

		s.handler.Update(c)

		s.Equal(http.StatusOK, w.Code)
	})

	s.Run(testCaseError_OrganizerUpdate_InvalidBody, func() {
		c, w := newOrganizerContext(http.MethodPut, "/organizer/events/e-1", []byte("{not-json"), gin.Params{{Key: "id", Value: "e-1"}})

		s.handler.Update(c)

		s.Equal(http.StatusBadRequest, w.Code)
	})
}

func (s *OrganizerEventHandlerTestSuite) TestDelete() {
	s.Run(testCaseSuccess_OrganizerDelete, func() {
		c, w := newOrganizerContext(http.MethodDelete, "/organizer/events/e-1", nil, gin.Params{{Key: "id", Value: "e-1"}})
		s.events.On("Cancel", mock.Anything, "e-1").Return(nil)

		s.handler.Delete(c)
		c.Writer.WriteHeaderNow() // c.Status() alone buffers until gin's engine flushes; force it for this direct call.

		s.Equal(http.StatusNoContent, w.Code)
	})

	s.Run(testCaseError_OrganizerDelete_ServiceError, func() {
		c, w := newOrganizerContext(http.MethodDelete, "/organizer/events/e-1", nil, gin.Params{{Key: "id", Value: "e-1"}})
		s.events.On("Cancel", mock.Anything, "e-1").Return(apperr.ErrConflict)

		s.handler.Delete(c)

		s.Equal(http.StatusConflict, w.Code)
	})
}

func (s *OrganizerEventHandlerTestSuite) TestPublish() {
	s.Run(testCaseSuccess_OrganizerPublish, func() {
		c, w := newOrganizerContext(http.MethodPost, "/organizer/events/e-1/publish", nil, gin.Params{{Key: "id", Value: "e-1"}})
		published := &model.Event{ID: "e-1", Status: model.StatusPublished}
		detail := &service.EventDetail{Event: published}
		s.events.On("Publish", mock.Anything, "e-1").Return(published, nil)
		s.events.On("GetDetailByID", mock.Anything, "e-1").Return(detail, nil)

		s.handler.Publish(c)

		s.Equal(http.StatusOK, w.Code)
	})

	s.Run(testCaseError_OrganizerPublish_PublishFails, func() {
		c, w := newOrganizerContext(http.MethodPost, "/organizer/events/e-1/publish", nil, gin.Params{{Key: "id", Value: "e-1"}})
		s.events.On("Publish", mock.Anything, "e-1").Return(nil, apperr.ErrConflict)

		s.handler.Publish(c)

		s.Equal(http.StatusConflict, w.Code)
	})

	s.Run(testCaseError_OrganizerPublish_GetDetailFails, func() {
		c, w := newOrganizerContext(http.MethodPost, "/organizer/events/e-1/publish", nil, gin.Params{{Key: "id", Value: "e-1"}})
		published := &model.Event{ID: "e-1", Status: model.StatusPublished}
		s.events.On("Publish", mock.Anything, "e-1").Return(published, nil)
		s.events.On("GetDetailByID", mock.Anything, "e-1").Return(nil, apperr.ErrInternal)

		s.handler.Publish(c)

		s.Equal(http.StatusInternalServerError, w.Code)
	})
}

func (s *OrganizerEventHandlerTestSuite) TestOwnerLookup() {
	s.Run(testCaseSuccess_OrganizerOwnerLookup, func() {
		c, _ := newOrganizerContext(http.MethodGet, "/organizer/events/e-1", nil, gin.Params{{Key: "id", Value: "e-1"}})
		s.events.On("GetOwnerID", mock.Anything, "e-1").Return("user-1", nil)

		owner, err := s.handler.OwnerLookup(c)

		s.Require().NoError(err)
		s.Equal("user-1", owner)
	})
}
