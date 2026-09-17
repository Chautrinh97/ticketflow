package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	gormpostgres "gorm.io/driver/postgres"

	"ticketflow/pkg/apperr"
	"ticketflow/pkg/pagination"
	"ticketflow/services/event-service/internal/model"
	"ticketflow/services/event-service/internal/repository"
	"ticketflow/services/event-service/internal/service/mocks"
)

// Case-name constants (tài liệu sống) per
// docs/01-architecture/backend-conventions.md's testing convention.
const (
	caseSuccessGetDetailBySlugFound                      = "[Success] Trả về chi tiết sự kiện khi slug tồn tại và đã publish"
	caseErrorGetDetailBySlugNotFound                     = "[Error] Trả về apperr.ErrNotFound khi slug không tồn tại"
	caseErrorGetDetailBySlugRepoError                    = "[Error] Trả về lỗi gốc khi repository lỗi khác gorm.ErrRecordNotFound"
	caseSuccessGetDetailBySlugCatalogNoDocumentsFallback = "[Success] Dùng catalog mặc định khi Mongo trả về mongo.ErrNoDocuments"
	caseErrorGetDetailBySlugCatalogError                 = "[Error] Trả về lỗi gốc khi catalog lỗi khác mongo.ErrNoDocuments"
	caseErrorGetDetailBySlugTicketTypesError             = "[Error] Trả về lỗi gốc khi ticket_types repo lỗi, không gọi catalog nữa"

	caseSuccessGetDetailByIDFound   = "[Success] Trả về chi tiết sự kiện khi id tồn tại"
	caseErrorGetDetailByIDNotFound  = "[Error] Trả về apperr.ErrNotFound khi id không tồn tại"
	caseErrorGetDetailByIDRepoError = "[Error] Trả về lỗi gốc khi repository lỗi khác gorm.ErrRecordNotFound"

	caseSuccessGetOwnerID        = "[Success] Trả về đúng organizer_id của sự kiện"
	caseErrorGetOwnerIDRepoError = "[Error] Trả về lỗi gốc khi repository lỗi"

	caseSuccessListPublished = "[Success] Chuyển tiếp filter/pagination và trả nguyên kết quả từ repository"

	caseSuccessListByOrganizer = "[Success] Chuyển tiếp organizerID/status/pagination và trả nguyên kết quả từ repository"

	caseSuccessPublishDraftToPublished = "[Success] Chuyển draft -> published"
	caseErrorPublishNotFound           = "[Error] Trả về apperr.ErrNotFound khi sự kiện không tồn tại"
	caseErrorPublishConflictNotDraft   = "[Error] Trả về apperr.ErrConflict khi sự kiện không ở trạng thái draft"
	caseErrorPublishUpdateFails        = "[Error] Trả về lỗi gốc khi repository.Update thất bại"

	caseErrorUpdateNotFound                     = "[Error] Trả về apperr.ErrNotFound khi sự kiện không tồn tại"
	caseSuccessUpdatePartialFieldsOnly          = "[Success] Chỉ merge field được truyền, giữ nguyên các field còn lại, không đụng tới catalog"
	caseSuccessUpdateAttributesAndTags          = "[Success] Attributes/Tags khác nil -> đọc rồi upsert lại catalog"
	caseSuccessUpdateCatalogNoDocumentsFallback = "[Success] Catalog chưa tồn tại (mongo.ErrNoDocuments) -> tạo mới rồi upsert"
	caseErrorUpdateEventsUpdateFails            = "[Error] Trả về lỗi gốc khi events.Update thất bại, không đụng tới catalog"
	caseErrorUpdateCatalogReadFails             = "[Error] Trả về lỗi gốc khi catalog.GetByEventID lỗi khác mongo.ErrNoDocuments, không upsert"
	caseErrorUpdateCatalogUpsertFails           = "[Error] Trả về lỗi gốc khi catalog.Upsert thất bại"

	caseSuccessCancel          = "[Success] Chuyển trạng thái sự kiện sang cancelled"
	caseErrorCancelNotFound    = "[Error] Trả về apperr.ErrNotFound khi sự kiện không tồn tại"
	caseErrorCancelUpdateFails = "[Error] Trả về lỗi gốc khi repository.Update thất bại"

	caseSuccessUniqueSlugNoCollision             = "[Success] Không trùng slug -> trả về slug gốc ngay lần thử đầu"
	caseSuccessUniqueSlugErrorTreatedAsNotExists = "[Success] SlugExists lỗi -> coi như không trùng, trả về slug hiện tại ngay"
	caseSuccessUniqueSlugOneCollisionThenSuccess = "[Success] Trùng 1 lần rồi không trùng -> trả về slug có hậu tố 8 ký tự"
	caseSuccessUniqueSlugRetryCapExhausted       = "[Success] Trùng cả 5 lần thử -> vượt retry cap, trả về slug có hậu tố UUID đầy đủ"
)

var errBoom = errors.New("boom: lỗi giả lập từ mock")

type EventServiceTestSuite struct {
	suite.Suite
}

func TestEventServiceTestSuite(t *testing.T) {
	suite.Run(t, new(EventServiceTestSuite))
}

// newSUT (system under test) wires a fresh EventService against 3 fresh
// mocks for every sub-test — mocks.New<X>(t) registers a t.Cleanup that
// calls AssertExpectations, so every .On(...) set up in a sub-test must
// actually be invoked by the code path under test.
func (s *EventServiceTestSuite) newSUT() (*EventService, *mocks.EventRepository, *mocks.TicketTypeRepository, *mocks.CatalogRepository) {
	eventsRepo := mocks.NewEventRepository(s.T())
	ticketTypesRepo := mocks.NewTicketTypeRepository(s.T())
	catalogRepo := mocks.NewCatalogRepository(s.T())
	return NewEventService(eventsRepo, ticketTypesRepo, catalogRepo), eventsRepo, ticketTypesRepo, catalogRepo
}

func (s *EventServiceTestSuite) requireAppErrCode(err error, want *apperr.Error) {
	s.T().Helper()
	var appErr *apperr.Error
	s.Require().ErrorAs(err, &appErr)
	s.Assert().Equal(want.Code, appErr.Code)
}

func sampleEvent() *model.Event {
	return &model.Event{
		ID:          "event-1",
		OrganizerID: "org-1",
		Title:       "Rock Concert",
		Slug:        "rock-concert",
		Category:    model.CategoryConcert,
		StartTime:   time.Date(2026, 1, 1, 19, 0, 0, 0, time.UTC),
		Status:      model.StatusDraft,
	}
}

func (s *EventServiceTestSuite) TestGetDetailBySlug() {
	ctx := context.Background()

	s.Run(caseSuccessGetDetailBySlugFound, func() {
		svc, eventsRepo, ticketTypesRepo, catalogRepo := s.newSUT()
		event := sampleEvent()
		event.Status = model.StatusPublished
		ticketTypes := []model.TicketType{{ID: "tt-1", EventID: event.ID, Price: 100000}}
		catalog := &model.EventCatalog{EventID: event.ID, Category: event.Category}

		eventsRepo.On("GetPublishedBySlug", ctx, "rock-concert").Return(event, nil)
		ticketTypesRepo.On("ListByEventID", ctx, event.ID).Return(ticketTypes, nil)
		catalogRepo.On("GetByEventID", ctx, event.ID).Return(catalog, nil)

		detail, err := svc.GetDetailBySlug(ctx, "rock-concert")

		s.Require().NoError(err)
		s.Assert().Same(event, detail.Event)
		s.Assert().Equal(ticketTypes, detail.TicketTypes)
		s.Assert().Same(catalog, detail.Catalog)
	})

	s.Run(caseErrorGetDetailBySlugNotFound, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		eventsRepo.On("GetPublishedBySlug", ctx, "missing").Return(nil, gorm.ErrRecordNotFound)

		detail, err := svc.GetDetailBySlug(ctx, "missing")

		s.requireAppErrCode(err, apperr.ErrNotFound)
		s.Assert().Nil(detail)
	})

	s.Run(caseErrorGetDetailBySlugRepoError, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		eventsRepo.On("GetPublishedBySlug", ctx, "rock-concert").Return(nil, errBoom)

		detail, err := svc.GetDetailBySlug(ctx, "rock-concert")

		s.Require().ErrorIs(err, errBoom)
		s.Assert().Nil(detail)
	})

	s.Run(caseSuccessGetDetailBySlugCatalogNoDocumentsFallback, func() {
		svc, eventsRepo, ticketTypesRepo, catalogRepo := s.newSUT()
		event := sampleEvent()
		event.Status = model.StatusPublished
		eventsRepo.On("GetPublishedBySlug", ctx, "rock-concert").Return(event, nil)
		ticketTypesRepo.On("ListByEventID", ctx, event.ID).Return(nil, nil)
		catalogRepo.On("GetByEventID", ctx, event.ID).Return(nil, mongo.ErrNoDocuments)

		detail, err := svc.GetDetailBySlug(ctx, "rock-concert")

		s.Require().NoError(err)
		s.Require().NotNil(detail.Catalog)
		s.Assert().Equal(&model.EventCatalog{EventID: event.ID, Category: event.Category}, detail.Catalog)
	})

	s.Run(caseErrorGetDetailBySlugCatalogError, func() {
		svc, eventsRepo, ticketTypesRepo, catalogRepo := s.newSUT()
		event := sampleEvent()
		event.Status = model.StatusPublished
		eventsRepo.On("GetPublishedBySlug", ctx, "rock-concert").Return(event, nil)
		ticketTypesRepo.On("ListByEventID", ctx, event.ID).Return(nil, nil)
		catalogRepo.On("GetByEventID", ctx, event.ID).Return(nil, errBoom)

		detail, err := svc.GetDetailBySlug(ctx, "rock-concert")

		s.Require().ErrorIs(err, errBoom)
		s.Assert().Nil(detail)
	})

	s.Run(caseErrorGetDetailBySlugTicketTypesError, func() {
		svc, eventsRepo, ticketTypesRepo, catalogRepo := s.newSUT()
		event := sampleEvent()
		event.Status = model.StatusPublished
		eventsRepo.On("GetPublishedBySlug", ctx, "rock-concert").Return(event, nil)
		ticketTypesRepo.On("ListByEventID", ctx, event.ID).Return(nil, errBoom)

		detail, err := svc.GetDetailBySlug(ctx, "rock-concert")

		s.Require().ErrorIs(err, errBoom)
		s.Assert().Nil(detail)
		catalogRepo.AssertNotCalled(s.T(), "GetByEventID", mock.Anything, mock.Anything)
	})
}

func (s *EventServiceTestSuite) TestGetDetailByID() {
	ctx := context.Background()

	s.Run(caseSuccessGetDetailByIDFound, func() {
		svc, eventsRepo, ticketTypesRepo, catalogRepo := s.newSUT()
		event := sampleEvent()
		ticketTypes := []model.TicketType{{ID: "tt-1", EventID: event.ID, Price: 100000}}
		catalog := &model.EventCatalog{EventID: event.ID, Category: event.Category}
		eventsRepo.On("GetByID", ctx, event.ID).Return(event, nil)
		ticketTypesRepo.On("ListByEventID", ctx, event.ID).Return(ticketTypes, nil)
		catalogRepo.On("GetByEventID", ctx, event.ID).Return(catalog, nil)

		detail, err := svc.GetDetailByID(ctx, event.ID)

		s.Require().NoError(err)
		s.Assert().Same(event, detail.Event)
	})

	s.Run(caseErrorGetDetailByIDNotFound, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		eventsRepo.On("GetByID", ctx, "missing").Return(nil, gorm.ErrRecordNotFound)

		detail, err := svc.GetDetailByID(ctx, "missing")

		s.requireAppErrCode(err, apperr.ErrNotFound)
		s.Assert().Nil(detail)
	})

	s.Run(caseErrorGetDetailByIDRepoError, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		eventsRepo.On("GetByID", ctx, "event-1").Return(nil, errBoom)

		detail, err := svc.GetDetailByID(ctx, "event-1")

		s.Require().ErrorIs(err, errBoom)
		s.Assert().Nil(detail)
	})
}

func (s *EventServiceTestSuite) TestGetOwnerID() {
	ctx := context.Background()

	s.Run(caseSuccessGetOwnerID, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		event := sampleEvent()
		eventsRepo.On("GetByID", ctx, event.ID).Return(event, nil)

		ownerID, err := svc.GetOwnerID(ctx, event.ID)

		s.Require().NoError(err)
		s.Assert().Equal(event.OrganizerID, ownerID)
	})

	s.Run(caseErrorGetOwnerIDRepoError, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		eventsRepo.On("GetByID", ctx, "event-1").Return(nil, errBoom)

		ownerID, err := svc.GetOwnerID(ctx, "event-1")

		s.Require().ErrorIs(err, errBoom)
		s.Assert().Empty(ownerID)
	})
}

func (s *EventServiceTestSuite) TestListPublished() {
	s.Run(caseSuccessListPublished, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		ctx := context.Background()
		filter := repository.EventFilter{Category: model.CategoryConcert}
		p := pagination.Params{Page: 2, PageSize: 10}
		rows := []repository.EventSummaryRow{{ID: "e-1"}}
		eventsRepo.On("ListPublished", ctx, filter, p).Return(rows, int64(1), nil)

		got, total, err := svc.ListPublished(ctx, filter, p)

		s.Require().NoError(err)
		s.Assert().Equal(rows, got)
		s.Assert().EqualValues(1, total)
	})
}

func (s *EventServiceTestSuite) TestListByOrganizer() {
	s.Run(caseSuccessListByOrganizer, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		ctx := context.Background()
		p := pagination.Params{Page: 1, PageSize: 20}
		rows := []repository.EventSummaryRow{{ID: "e-1"}, {ID: "e-2"}}
		eventsRepo.On("ListByOrganizer", ctx, "org-1", "draft", p).Return(rows, int64(2), nil)

		got, total, err := svc.ListByOrganizer(ctx, "org-1", "draft", p)

		s.Require().NoError(err)
		s.Assert().Equal(rows, got)
		s.Assert().EqualValues(2, total)
	})
}

func (s *EventServiceTestSuite) TestPublish() {
	ctx := context.Background()

	s.Run(caseSuccessPublishDraftToPublished, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		event := sampleEvent()
		eventsRepo.On("GetByID", ctx, event.ID).Return(event, nil)
		eventsRepo.On("Update", ctx, mock.MatchedBy(func(e *model.Event) bool {
			return e.Status == model.StatusPublished
		})).Return(nil)

		got, err := svc.Publish(ctx, event.ID)

		s.Require().NoError(err)
		s.Assert().Equal(model.StatusPublished, got.Status)
	})

	s.Run(caseErrorPublishNotFound, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		eventsRepo.On("GetByID", ctx, "missing").Return(nil, gorm.ErrRecordNotFound)

		got, err := svc.Publish(ctx, "missing")

		s.requireAppErrCode(err, apperr.ErrNotFound)
		s.Assert().Nil(got)
	})

	s.Run(caseErrorPublishConflictNotDraft, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		event := sampleEvent()
		event.Status = model.StatusPublished
		eventsRepo.On("GetByID", ctx, event.ID).Return(event, nil)

		got, err := svc.Publish(ctx, event.ID)

		s.requireAppErrCode(err, apperr.ErrConflict)
		s.Assert().Nil(got)
		eventsRepo.AssertNotCalled(s.T(), "Update", mock.Anything, mock.Anything)
	})

	s.Run(caseErrorPublishUpdateFails, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		event := sampleEvent()
		eventsRepo.On("GetByID", ctx, event.ID).Return(event, nil)
		eventsRepo.On("Update", ctx, mock.Anything).Return(errBoom)

		got, err := svc.Publish(ctx, event.ID)

		s.Require().ErrorIs(err, errBoom)
		s.Assert().Nil(got)
	})
}

func (s *EventServiceTestSuite) TestUpdate() {
	ctx := context.Background()

	s.Run(caseErrorUpdateNotFound, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		eventsRepo.On("GetByID", ctx, "missing").Return(nil, gorm.ErrRecordNotFound)

		detail, err := svc.Update(ctx, "missing", UpdateEventInput{})

		s.requireAppErrCode(err, apperr.ErrNotFound)
		s.Assert().Nil(detail)
	})

	s.Run(caseSuccessUpdatePartialFieldsOnly, func() {
		svc, eventsRepo, ticketTypesRepo, catalogRepo := s.newSUT()
		event := sampleEvent()
		originalVenue := event.VenueName
		newTitle := "New Title"
		eventsRepo.On("GetByID", ctx, event.ID).Return(event, nil)
		eventsRepo.On("Update", ctx, mock.MatchedBy(func(e *model.Event) bool {
			return e.Title == newTitle && e.VenueName == originalVenue
		})).Return(nil)
		ticketTypesRepo.On("ListByEventID", ctx, event.ID).Return(nil, nil)
		catalogRepo.On("GetByEventID", ctx, event.ID).Return(&model.EventCatalog{EventID: event.ID}, nil)

		detail, err := svc.Update(ctx, event.ID, UpdateEventInput{Title: &newTitle})

		s.Require().NoError(err)
		s.Assert().Equal(newTitle, detail.Event.Title)
		catalogRepo.AssertNotCalled(s.T(), "Upsert", mock.Anything, mock.Anything)
		catalogRepo.AssertNumberOfCalls(s.T(), "GetByEventID", 1) // chỉ gọi ở assembleDetail cuối, không gọi để merge
	})

	s.Run(caseSuccessUpdateAttributesAndTags, func() {
		svc, eventsRepo, ticketTypesRepo, catalogRepo := s.newSUT()
		event := sampleEvent()
		newAttrs := map[string]any{"artists": []string{"Band A"}}
		newTags := []string{"rock"}
		existingCatalog := &model.EventCatalog{EventID: event.ID, Category: event.Category, Tags: []string{"old"}}
		eventsRepo.On("GetByID", ctx, event.ID).Return(event, nil)
		eventsRepo.On("Update", ctx, mock.Anything).Return(nil)
		catalogRepo.On("GetByEventID", ctx, event.ID).Return(existingCatalog, nil)
		catalogRepo.On("Upsert", ctx, mock.MatchedBy(func(c *model.EventCatalog) bool {
			return c.EventID == event.ID && len(c.Tags) == 1 && c.Tags[0] == "rock"
		})).Return(nil)
		ticketTypesRepo.On("ListByEventID", ctx, event.ID).Return(nil, nil)

		detail, err := svc.Update(ctx, event.ID, UpdateEventInput{Attributes: newAttrs, Tags: newTags})

		s.Require().NoError(err)
		s.Require().NotNil(detail)
		catalogRepo.AssertNumberOfCalls(s.T(), "GetByEventID", 2) // 1 lần để merge, 1 lần trong assembleDetail cuối
	})

	s.Run(caseSuccessUpdateCatalogNoDocumentsFallback, func() {
		svc, eventsRepo, ticketTypesRepo, catalogRepo := s.newSUT()
		event := sampleEvent()
		newTags := []string{"new-tag"}
		eventsRepo.On("GetByID", ctx, event.ID).Return(event, nil)
		eventsRepo.On("Update", ctx, mock.Anything).Return(nil)
		catalogRepo.On("GetByEventID", ctx, event.ID).Return(nil, mongo.ErrNoDocuments)
		catalogRepo.On("Upsert", ctx, mock.MatchedBy(func(c *model.EventCatalog) bool {
			return c.EventID == event.ID && c.Category == event.Category && len(c.Tags) == 1 && c.Tags[0] == "new-tag"
		})).Return(nil)
		ticketTypesRepo.On("ListByEventID", ctx, event.ID).Return(nil, nil)

		detail, err := svc.Update(ctx, event.ID, UpdateEventInput{Tags: newTags})

		s.Require().NoError(err)
		s.Require().NotNil(detail)
	})

	s.Run(caseErrorUpdateEventsUpdateFails, func() {
		svc, eventsRepo, _, catalogRepo := s.newSUT()
		event := sampleEvent()
		eventsRepo.On("GetByID", ctx, event.ID).Return(event, nil)
		eventsRepo.On("Update", ctx, mock.Anything).Return(errBoom)

		detail, err := svc.Update(ctx, event.ID, UpdateEventInput{Tags: []string{"x"}})

		s.Require().ErrorIs(err, errBoom)
		s.Assert().Nil(detail)
		catalogRepo.AssertNotCalled(s.T(), "GetByEventID", mock.Anything, mock.Anything)
	})

	s.Run(caseErrorUpdateCatalogReadFails, func() {
		svc, eventsRepo, ticketTypesRepo, catalogRepo := s.newSUT()
		event := sampleEvent()
		eventsRepo.On("GetByID", ctx, event.ID).Return(event, nil)
		eventsRepo.On("Update", ctx, mock.Anything).Return(nil)
		catalogRepo.On("GetByEventID", ctx, event.ID).Return(nil, errBoom)

		detail, err := svc.Update(ctx, event.ID, UpdateEventInput{Tags: []string{"x"}})

		s.Require().ErrorIs(err, errBoom)
		s.Assert().Nil(detail)
		catalogRepo.AssertNotCalled(s.T(), "Upsert", mock.Anything, mock.Anything)
		ticketTypesRepo.AssertNotCalled(s.T(), "ListByEventID", mock.Anything, mock.Anything)
	})

	s.Run(caseErrorUpdateCatalogUpsertFails, func() {
		svc, eventsRepo, ticketTypesRepo, catalogRepo := s.newSUT()
		event := sampleEvent()
		eventsRepo.On("GetByID", ctx, event.ID).Return(event, nil)
		eventsRepo.On("Update", ctx, mock.Anything).Return(nil)
		catalogRepo.On("GetByEventID", ctx, event.ID).Return(&model.EventCatalog{EventID: event.ID}, nil)
		catalogRepo.On("Upsert", ctx, mock.Anything).Return(errBoom)

		detail, err := svc.Update(ctx, event.ID, UpdateEventInput{Tags: []string{"x"}})

		s.Require().ErrorIs(err, errBoom)
		s.Assert().Nil(detail)
		ticketTypesRepo.AssertNotCalled(s.T(), "ListByEventID", mock.Anything, mock.Anything)
	})
}

func (s *EventServiceTestSuite) TestCancel() {
	ctx := context.Background()

	s.Run(caseSuccessCancel, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		event := sampleEvent()
		eventsRepo.On("GetByID", ctx, event.ID).Return(event, nil)
		eventsRepo.On("Update", ctx, mock.MatchedBy(func(e *model.Event) bool {
			return e.Status == model.StatusCancelled
		})).Return(nil)

		err := svc.Cancel(ctx, event.ID)

		s.Require().NoError(err)
	})

	s.Run(caseErrorCancelNotFound, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		eventsRepo.On("GetByID", ctx, "missing").Return(nil, gorm.ErrRecordNotFound)

		err := svc.Cancel(ctx, "missing")

		s.requireAppErrCode(err, apperr.ErrNotFound)
	})

	s.Run(caseErrorCancelUpdateFails, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		event := sampleEvent()
		eventsRepo.On("GetByID", ctx, event.ID).Return(event, nil)
		eventsRepo.On("Update", ctx, mock.Anything).Return(errBoom)

		err := svc.Cancel(ctx, event.ID)

		s.Require().ErrorIs(err, errBoom)
	})
}

var uuid8HexSuffix = regexp.MustCompile(`^[0-9a-f]{8}$`)
var fullUUIDSuffix = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func (s *EventServiceTestSuite) TestUniqueSlug() {
	ctx := context.Background()
	const base = "rock-concert"

	s.Run(caseSuccessUniqueSlugNoCollision, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		eventsRepo.On("SlugExists", ctx, base).Return(false, nil).Once()

		got := svc.uniqueSlug(ctx, "Rock Concert")

		s.Assert().Equal(base, got)
	})

	s.Run(caseSuccessUniqueSlugErrorTreatedAsNotExists, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		eventsRepo.On("SlugExists", ctx, base).Return(true, errBoom).Once()

		got := svc.uniqueSlug(ctx, "Rock Concert")

		s.Assert().Equal(base, got, "lỗi từ SlugExists phải được coi như 'không trùng' và trả về slug hiện tại ngay")
	})

	s.Run(caseSuccessUniqueSlugOneCollisionThenSuccess, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		eventsRepo.On("SlugExists", ctx, base).Return(true, nil).Once()
		eventsRepo.On("SlugExists", ctx, mock.MatchedBy(func(slug string) bool {
			return strings.HasPrefix(slug, base+"-")
		})).Return(false, nil).Once()

		got := svc.uniqueSlug(ctx, "Rock Concert")

		suffix := strings.TrimPrefix(got, base+"-")
		s.Require().True(uuid8HexSuffix.MatchString(suffix), "slug = %q, muốn hậu tố 8 ký tự hex", got)
		eventsRepo.AssertNumberOfCalls(s.T(), "SlugExists", 2)
	})

	s.Run(caseSuccessUniqueSlugRetryCapExhausted, func() {
		svc, eventsRepo, _, _ := s.newSUT()
		eventsRepo.On("SlugExists", ctx, mock.Anything).Return(true, nil).Times(5)

		got := svc.uniqueSlug(ctx, "Rock Concert")

		suffix := strings.TrimPrefix(got, base+"-")
		s.Require().True(fullUUIDSuffix.MatchString(suffix), "slug = %q, muốn hậu tố UUID đầy đủ sau khi vượt retry cap", got)
		eventsRepo.AssertNumberOfCalls(s.T(), "SlugExists", 5)
	})
}

// --- Create: cần Postgres thật ---
//
// EventService.Create luôn đi qua s.events.DB().WithContext(ctx).Transaction(...)
// (event_service.go) — GORM's Transaction() gọi thẳng db.Begin() trên kết nối
// thật, không có cách nào mock EventRepository.DB() để trả về 1 *gorm.DB
// "giả" nhưng vẫn chạy Transaction() được mà không có driver thật đứng sau
// (repo này không có go-sqlmock/sqlite trong danh sách dependency được phép
// dùng). Vì vậy nhánh "Mongo thất bại -> rollback Postgres" cần 1 Postgres
// thật (testcontainers), không thể test thuần bằng mock như các method khác
// ở trên — EventRepository dùng ở đây là *repository.EventPostgresRepo thật,
// chỉ CatalogRepository (đại diện cho Mongo) được mock để giả lập
// thành công/thất bại một cách xác định.

var (
	createTestGormDB        *gorm.DB
	createTestPostgresReady bool
)

func TestMain(m *testing.M) {
	os.Exit(runEventServiceTestMain(m))
}

func runEventServiceTestMain(m *testing.M) int {
	ctx := context.Background()

	pgC, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("ticketflow_test"),
		tcpostgres.WithUsername("ticketflow"),
		tcpostgres.WithPassword("ticketflow"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Printf("WARN: internal/service: không khởi được Postgres container (%v) — chỉ TestEventServiceCreate_Integration bị skip, EventServiceTestSuite (mock-based) ở trên vẫn chạy thật", err)
		return m.Run()
	}
	defer func() { _ = pgC.Terminate(ctx) }()

	connStr, err := pgC.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("WARN: internal/service: không lấy được postgres connection string: %v", err)
		return m.Run()
	}

	if err := seedUsersStubTable(connStr); err != nil {
		log.Printf("WARN: internal/service: không tạo được bảng users stub: %v", err)
		return m.Run()
	}
	if err := applyEventServiceMigrations(connStr); err != nil {
		log.Printf("WARN: internal/service: không apply được migrations: %v", err)
		return m.Run()
	}

	gormDB, err := gorm.Open(gormpostgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Printf("WARN: internal/service: không mở được kết nối gorm: %v", err)
		return m.Run()
	}

	createTestGormDB = gormDB
	createTestPostgresReady = true
	return m.Run()
}

// seedUsersStubTable tạo phần tối thiểu của bảng `users` (thuộc
// identity-service) mà FK events.organizer_id cần — event-service không
// chạy migration của identity-service ở đây, cùng cách tiếp cận với
// booking-service/internal/repository/booking_repository_concurrency_test.go
// cho cùng vấn đề FK xuyên service.
func seedUsersStubTable(connStr string) error {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto; CREATE TABLE IF NOT EXISTS users (id UUID PRIMARY KEY DEFAULT gen_random_uuid())`)
	return err
}

// applyEventServiceMigrations chạy thật migrations/*.up.sql của event-service
// qua golang-migrate, đúng convention ở backend-conventions.md — không
// hand-copy schema events/ticket_types.
func applyEventServiceMigrations(connStr string) error {
	mig, err := migrate.New("file://../../migrations", connStr)
	if err != nil {
		return err
	}
	defer func() { _, _ = mig.Close() }()
	if err := mig.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func seedOrganizerUser(t *testing.T, db *gorm.DB) string {
	t.Helper()
	id := uuid.NewString()
	require.NoError(t, db.Exec("INSERT INTO users (id) VALUES (?)", id).Error)
	return id
}

func truncateEventAndTicketTypeTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec("TRUNCATE ticket_types, events").Error)
}

func TestEventServiceCreate_Integration(t *testing.T) {
	if !createTestPostgresReady {
		t.Skip("skip: không có Postgres thật trong môi trường này (xem cảnh báo ở TestMain) — EventService.Create bọc 1 transaction *gorm.DB thật (DB().Transaction(...)), mock không thể giả lập; cần Postgres thật, khác với phần còn lại của EventServiceTestSuite")
	}

	eventsRepo := repository.NewEventPostgresRepo(createTestGormDB)
	ticketTypesRepo := repository.NewTicketTypePostgresRepo(createTestGormDB)
	organizerID := seedOrganizerUser(t, createTestGormDB)

	t.Run("[Success] Mongo upsert thành công -> Postgres commit thật sự", func(t *testing.T) {
		t.Cleanup(func() { truncateEventAndTicketTypeTables(t, createTestGormDB) })
		catalogRepo := mocks.NewCatalogRepository(t)
		catalogRepo.On("Upsert", mock.Anything, mock.Anything).Return(nil)
		svc := NewEventService(eventsRepo, ticketTypesRepo, catalogRepo)

		detail, err := svc.Create(context.Background(), organizerID, CreateEventInput{
			Title: "Integration Concert", Category: model.CategoryConcert, StartTime: time.Now().Add(24 * time.Hour),
		})

		require.NoError(t, err)
		require.NotNil(t, detail)

		var count int64
		require.NoError(t, createTestGormDB.Model(&model.Event{}).Where("id = ?", detail.Event.ID).Count(&count).Error)
		assert.EqualValues(t, 1, count, "event row phải được commit thật vào Postgres")
	})

	t.Run("[Error] Mongo upsert thất bại -> rollback, không để lại events mồ côi", func(t *testing.T) {
		t.Cleanup(func() { truncateEventAndTicketTypeTables(t, createTestGormDB) })
		catalogRepo := mocks.NewCatalogRepository(t)
		catalogRepo.On("Upsert", mock.Anything, mock.Anything).Return(fmt.Errorf("mongo write failed"))
		svc := NewEventService(eventsRepo, ticketTypesRepo, catalogRepo)

		detail, err := svc.Create(context.Background(), organizerID, CreateEventInput{
			Title: "Rollback Concert", Category: model.CategoryConcert, StartTime: time.Now().Add(24 * time.Hour),
		})

		require.Error(t, err)
		assert.Nil(t, detail)

		var count int64
		require.NoError(t, createTestGormDB.Model(&model.Event{}).Where("title = ?", "Rollback Concert").Count(&count).Error)
		assert.EqualValues(t, 0, count, "không được để lại events row mồ côi sau khi Mongo write thất bại")
	})
}
