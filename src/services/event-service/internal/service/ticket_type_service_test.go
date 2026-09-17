package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"ticketflow/pkg/apperr"
	"ticketflow/services/event-service/internal/model"
	"ticketflow/services/event-service/internal/service/mocks"
)

const (
	caseErrorTicketTypeCreateEventNotFound      = "[Error] Trả về apperr.ErrNotFound khi sự kiện không tồn tại"
	caseErrorTicketTypeCreateEventLookupFails   = "[Error] Trả về lỗi gốc khi kiểm tra sự kiện lỗi khác gorm.ErrRecordNotFound"
	caseSuccessTicketTypeCreateDefaultCurrency  = "[Success] Currency rỗng -> mặc định VND"
	caseSuccessTicketTypeCreateExplicitCurrency = "[Success] Currency có giá trị -> giữ nguyên, không ghi đè"
	caseErrorTicketTypeCreateWriteFails         = "[Error] Trả về lỗi gốc khi ticketTypes.Create thất bại"
)

type TicketTypeServiceTestSuite struct {
	suite.Suite
}

func TestTicketTypeServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TicketTypeServiceTestSuite))
}

func (s *TicketTypeServiceTestSuite) newSUT() (*TicketTypeService, *mocks.EventLookup, *mocks.TicketTypeWriter) {
	eventLookup := mocks.NewEventLookup(s.T())
	writer := mocks.NewTicketTypeWriter(s.T())
	return NewTicketTypeService(eventLookup, writer), eventLookup, writer
}

func (s *TicketTypeServiceTestSuite) TestCreate() {
	ctx := context.Background()
	const eventID = "event-1"

	s.Run(caseErrorTicketTypeCreateEventNotFound, func() {
		svc, eventLookup, writer := s.newSUT()
		eventLookup.On("GetByID", ctx, eventID).Return(nil, gorm.ErrRecordNotFound)

		tt, err := svc.Create(ctx, eventID, CreateTicketTypeInput{Name: "VIP", Price: 100000, Quota: 10})

		var appErr *apperr.Error
		s.Require().ErrorAs(err, &appErr)
		s.Assert().Equal(apperr.ErrNotFound.Code, appErr.Code)
		s.Assert().Nil(tt)
		writer.AssertNotCalled(s.T(), "Create", mock.Anything, mock.Anything)
	})

	s.Run(caseErrorTicketTypeCreateEventLookupFails, func() {
		svc, eventLookup, _ := s.newSUT()
		eventLookup.On("GetByID", ctx, eventID).Return(nil, errBoom)

		tt, err := svc.Create(ctx, eventID, CreateTicketTypeInput{Name: "VIP", Price: 100000, Quota: 10})

		s.Require().ErrorIs(err, errBoom)
		s.Assert().Nil(tt)
	})

	s.Run(caseSuccessTicketTypeCreateDefaultCurrency, func() {
		svc, eventLookup, writer := s.newSUT()
		eventLookup.On("GetByID", ctx, eventID).Return(&model.Event{ID: eventID}, nil)
		writer.On("Create", ctx, mock.MatchedBy(func(tt *model.TicketType) bool {
			return tt.Currency == "VND" && tt.EventID == eventID
		})).Return(nil)

		tt, err := svc.Create(ctx, eventID, CreateTicketTypeInput{Name: "VIP", Price: 100000, Quota: 10, Currency: ""})

		s.Require().NoError(err)
		s.Require().NotNil(tt)
		s.Assert().Equal("VND", tt.Currency)
	})

	s.Run(caseSuccessTicketTypeCreateExplicitCurrency, func() {
		svc, eventLookup, writer := s.newSUT()
		eventLookup.On("GetByID", ctx, eventID).Return(&model.Event{ID: eventID}, nil)
		writer.On("Create", ctx, mock.MatchedBy(func(tt *model.TicketType) bool {
			return tt.Currency == "USD"
		})).Return(nil)

		tt, err := svc.Create(ctx, eventID, CreateTicketTypeInput{Name: "VIP", Price: 100000, Quota: 10, Currency: "USD"})

		s.Require().NoError(err)
		s.Assert().Equal("USD", tt.Currency)
	})

	s.Run(caseErrorTicketTypeCreateWriteFails, func() {
		svc, eventLookup, writer := s.newSUT()
		eventLookup.On("GetByID", ctx, eventID).Return(&model.Event{ID: eventID}, nil)
		writer.On("Create", ctx, mock.Anything).Return(errBoom)

		tt, err := svc.Create(ctx, eventID, CreateTicketTypeInput{Name: "VIP", Price: 100000, Quota: 10})

		s.Require().ErrorIs(err, errBoom)
		s.Assert().Nil(tt)
	})
}
