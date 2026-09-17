package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"ticketflow/pkg/apperr"
	"ticketflow/pkg/pagination"
	"ticketflow/services/booking-service/internal/eventclient"
	"ticketflow/services/booking-service/internal/lock"
	"ticketflow/services/booking-service/internal/model"
	"ticketflow/services/booking-service/internal/service/mocks"
)

// Test-case name constants — [Success]/[Error] prefix documents expected
// outcome, per backend-conventions.md "Cấu trúc test".
const (
	testCaseError_CreateOrder_EmptyItems              = "[Error] Từ chối khi items rỗng"
	testCaseError_CreateOrder_InvalidQuantity         = "[Error] Từ chối khi quantity < 1"
	testCaseSuccess_CreateOrder_DedupsTicketTypeIDs   = "[Success] Gộp trùng ticket_type_id trước khi lock/query"
	testCaseError_CreateOrder_LockFails               = "[Error] Trả lỗi khi acquire lock thất bại, không gọi repo/event"
	testCaseError_CreateOrder_GetEventIDsFails        = "[Error] Lan truyền lỗi từ GetEventIDsForTicketTypes, vẫn unlock"
	testCaseError_CreateOrder_TicketTypeNotFound      = "[Error] Một số ticket_type không tồn tại -> NotFound"
	testCaseError_CreateOrder_EventNotPublished       = "[Error] Từ chối khi event chưa published, không gọi repo.CreateOrder"
	testCaseError_CreateOrder_EventClientError        = "[Error] Lan truyền lỗi gRPC từ EventChecker qua FromGRPCError"
	testCaseError_CreateOrder_RepoCreateOrderConflict = "[Error] Lan truyền lỗi conflict từ repo.CreateOrder"
	testCaseSuccess_CreateOrder_HappyPath             = "[Success] Trả về order khi mọi bước đều thành công"
	testCaseSuccess_CreateOrder_CallOrder             = "[Success] Thứ tự gọi: lock -> resolve event -> check published -> repo.CreateOrder"

	testCaseSuccess_ConfirmPayment_Passthrough = "[Success] Uỷ quyền thẳng cho repo.ConfirmPayment"
	testCaseError_ConfirmPayment_RepoError     = "[Error] Lan truyền lỗi từ repo.ConfirmPayment"

	testCaseSuccess_FailPayment_Passthrough = "[Success] Uỷ quyền thẳng cho repo.FailPayment"
	testCaseError_FailPayment_RepoError     = "[Error] Lan truyền lỗi từ repo.FailPayment"

	testCaseSuccess_GetOwnerID_ReturnsUserID = "[Success] Trả về order.UserID"
	testCaseError_GetOwnerID_RepoError       = "[Error] Lan truyền lỗi khi không tìm thấy order"

	testCaseSuccess_ListMyOrders_PassesLimitOffset = "[Success] Chuyển page/page_size thành limit/offset đúng"
	testCaseError_ListMyOrders_RepoError           = "[Error] Lan truyền lỗi từ repo.ListOrdersByUser"

	testCaseSuccess_GetOrderByID_Passthrough = "[Success] Uỷ quyền thẳng cho repo.GetOrderByID"
)

type BookingServiceTestSuite struct {
	suite.Suite
}

func TestBookingServiceTestSuite(t *testing.T) {
	suite.Run(t, new(BookingServiceTestSuite))
}

func noopUnlock() {}

// deps is a fresh set of mocked dependencies + the service under test.
// Built per sub-test (not once in SetupTest) so `.On(...)` expectations
// from one `s.Run` case never bleed into another — testify's SetupTest
// runs once per Test method, not per s.Run sub-test.
type deps struct {
	repo    *mocks.Repository
	locker  *mocks.Locker
	eventCl *mocks.EventChecker
	svc     *BookingService
}

func (s *BookingServiceTestSuite) newDeps() *deps {
	d := &deps{
		repo:    mocks.NewRepository(s.T()),
		locker:  mocks.NewLocker(s.T()),
		eventCl: mocks.NewEventChecker(s.T()),
	}
	d.svc = NewBookingService(d.repo, d.locker, d.eventCl)
	return d
}

func (s *BookingServiceTestSuite) TestCreateOrder() {
	s.Run(testCaseError_CreateOrder_EmptyItems, func() {
		d := s.newDeps()
		order, err := d.svc.CreateOrder(context.Background(), "user-1", nil)
		s.Require().Nil(order)
		s.Require().True(isAppErrCode(err, apperr.ErrValidation.Code))
		// No dependency should be touched at all for a request rejected
		// before any lock/repo/event interaction.
		d.repo.AssertNotCalled(s.T(), "GetEventIDsForTicketTypes", mock.Anything, mock.Anything)
		d.locker.AssertNotCalled(s.T(), "AcquireMany", mock.Anything, mock.Anything)
	})

	s.Run(testCaseError_CreateOrder_InvalidQuantity, func() {
		d := s.newDeps()
		items := []model.BookingItem{{TicketTypeID: "tt-1", Quantity: 0}}
		order, err := d.svc.CreateOrder(context.Background(), "user-1", items)
		s.Require().Nil(order)
		s.Require().True(isAppErrCode(err, apperr.ErrValidation.Code))
		d.locker.AssertNotCalled(s.T(), "AcquireMany", mock.Anything, mock.Anything)
	})

	s.Run(testCaseSuccess_CreateOrder_DedupsTicketTypeIDs, func() {
		d := s.newDeps()
		items := []model.BookingItem{
			{TicketTypeID: "tt-1", Quantity: 2},
			{TicketTypeID: "tt-1", Quantity: 3},
			{TicketTypeID: "tt-2", Quantity: 1},
		}
		var lockedIDs []string
		d.locker.On("AcquireMany", mock.Anything, mock.MatchedBy(func(ids []string) bool {
			lockedIDs = ids
			return true
		})).Return(lock.Unlock(noopUnlock), nil)
		d.repo.On("GetEventIDsForTicketTypes", mock.Anything, mock.Anything).
			Return(map[string]string{"tt-1": "ev-1", "tt-2": "ev-1"}, nil)
		d.eventCl.On("GetEvent", mock.Anything, "ev-1").
			Return(&eventclient.Event{ID: "ev-1", Status: "published"}, nil)
		wantOrder := &model.Order{ID: "order-1"}
		d.repo.On("CreateOrder", mock.Anything, "user-1", items).Return(wantOrder, nil)

		order, err := d.svc.CreateOrder(context.Background(), "user-1", items)
		s.Require().NoError(err)
		s.Require().Same(wantOrder, order)
		s.Assert().ElementsMatch([]string{"tt-1", "tt-2"}, lockedIDs)
		s.Assert().Len(lockedIDs, 2, "tt-1 must be deduped to a single lock entry despite appearing twice in items")
	})

	s.Run(testCaseError_CreateOrder_LockFails, func() {
		d := s.newDeps()
		items := []model.BookingItem{{TicketTypeID: "tt-1", Quantity: 1}}
		lockErr := apperr.WithMessage(apperr.ErrTooManyRequests, "busy")
		d.locker.On("AcquireMany", mock.Anything, []string{"tt-1"}).Return(nil, lockErr)

		order, err := d.svc.CreateOrder(context.Background(), "user-1", items)
		s.Require().Nil(order)
		s.Require().Same(lockErr, err)
		d.repo.AssertNotCalled(s.T(), "GetEventIDsForTicketTypes", mock.Anything, mock.Anything)
		d.repo.AssertNotCalled(s.T(), "CreateOrder", mock.Anything, mock.Anything, mock.Anything)
		d.eventCl.AssertNotCalled(s.T(), "GetEvent", mock.Anything, mock.Anything)
	})

	s.Run(testCaseError_CreateOrder_GetEventIDsFails, func() {
		d := s.newDeps()
		items := []model.BookingItem{{TicketTypeID: "tt-1", Quantity: 1}}
		unlockCalled := false
		d.locker.On("AcquireMany", mock.Anything, []string{"tt-1"}).
			Return(lock.Unlock(func() { unlockCalled = true }), nil)
		repoErr := errors.New("db down")
		d.repo.On("GetEventIDsForTicketTypes", mock.Anything, []string{"tt-1"}).Return(nil, repoErr)

		order, err := d.svc.CreateOrder(context.Background(), "user-1", items)
		s.Require().Nil(order)
		s.Require().Same(repoErr, err)
		s.Assert().True(unlockCalled, "lock must be released even when a later step fails")
		d.eventCl.AssertNotCalled(s.T(), "GetEvent", mock.Anything, mock.Anything)
		d.repo.AssertNotCalled(s.T(), "CreateOrder", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run(testCaseError_CreateOrder_TicketTypeNotFound, func() {
		d := s.newDeps()
		items := []model.BookingItem{
			{TicketTypeID: "tt-1", Quantity: 1},
			{TicketTypeID: "tt-missing", Quantity: 1},
		}
		d.locker.On("AcquireMany", mock.Anything, mock.Anything).Return(lock.Unlock(noopUnlock), nil)
		// Only tt-1 resolves to an event -> length mismatch vs 2 requested ids.
		d.repo.On("GetEventIDsForTicketTypes", mock.Anything, mock.Anything).
			Return(map[string]string{"tt-1": "ev-1"}, nil)

		order, err := d.svc.CreateOrder(context.Background(), "user-1", items)
		s.Require().Nil(order)
		s.Require().True(isAppErrCode(err, apperr.ErrNotFound.Code))
		d.eventCl.AssertNotCalled(s.T(), "GetEvent", mock.Anything, mock.Anything)
		d.repo.AssertNotCalled(s.T(), "CreateOrder", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run(testCaseError_CreateOrder_EventNotPublished, func() {
		d := s.newDeps()
		items := []model.BookingItem{{TicketTypeID: "tt-1", Quantity: 1}}
		d.locker.On("AcquireMany", mock.Anything, mock.Anything).Return(lock.Unlock(noopUnlock), nil)
		d.repo.On("GetEventIDsForTicketTypes", mock.Anything, mock.Anything).
			Return(map[string]string{"tt-1": "ev-1"}, nil)
		d.eventCl.On("GetEvent", mock.Anything, "ev-1").
			Return(&eventclient.Event{ID: "ev-1", Status: "draft"}, nil)

		order, err := d.svc.CreateOrder(context.Background(), "user-1", items)
		s.Require().Nil(order)
		s.Require().True(isAppErrCode(err, apperr.ErrConflict.Code))
		d.repo.AssertNotCalled(s.T(), "CreateOrder", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run(testCaseError_CreateOrder_EventClientError, func() {
		d := s.newDeps()
		items := []model.BookingItem{{TicketTypeID: "tt-1", Quantity: 1}}
		d.locker.On("AcquireMany", mock.Anything, mock.Anything).Return(lock.Unlock(noopUnlock), nil)
		d.repo.On("GetEventIDsForTicketTypes", mock.Anything, mock.Anything).
			Return(map[string]string{"tt-1": "ev-1"}, nil)
		d.eventCl.On("GetEvent", mock.Anything, "ev-1").
			Return(nil, errors.New("not a grpc status error"))

		order, err := d.svc.CreateOrder(context.Background(), "user-1", items)
		s.Require().Nil(order)
		// FromGRPCError falls back to ErrInternal for a non-status error.
		s.Require().True(isAppErrCode(err, apperr.ErrInternal.Code))
		d.repo.AssertNotCalled(s.T(), "CreateOrder", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run(testCaseError_CreateOrder_RepoCreateOrderConflict, func() {
		d := s.newDeps()
		items := []model.BookingItem{{TicketTypeID: "tt-1", Quantity: 1}}
		d.locker.On("AcquireMany", mock.Anything, mock.Anything).Return(lock.Unlock(noopUnlock), nil)
		d.repo.On("GetEventIDsForTicketTypes", mock.Anything, mock.Anything).
			Return(map[string]string{"tt-1": "ev-1"}, nil)
		d.eventCl.On("GetEvent", mock.Anything, "ev-1").
			Return(&eventclient.Event{ID: "ev-1", Status: "published"}, nil)
		conflictErr := apperr.WithMessage(apperr.ErrConflict, "Không đủ tồn kho")
		d.repo.On("CreateOrder", mock.Anything, "user-1", items).Return(nil, conflictErr)

		order, err := d.svc.CreateOrder(context.Background(), "user-1", items)
		s.Require().Nil(order)
		s.Require().Same(conflictErr, err)
	})

	s.Run(testCaseSuccess_CreateOrder_HappyPath, func() {
		d := s.newDeps()
		items := []model.BookingItem{{TicketTypeID: "tt-1", Quantity: 1}}
		d.locker.On("AcquireMany", mock.Anything, mock.Anything).Return(lock.Unlock(noopUnlock), nil)
		d.repo.On("GetEventIDsForTicketTypes", mock.Anything, mock.Anything).
			Return(map[string]string{"tt-1": "ev-1"}, nil)
		d.eventCl.On("GetEvent", mock.Anything, "ev-1").
			Return(&eventclient.Event{ID: "ev-1", Status: "published"}, nil)
		wantOrder := &model.Order{ID: "order-1", UserID: "user-1"}
		d.repo.On("CreateOrder", mock.Anything, "user-1", items).Return(wantOrder, nil)

		order, err := d.svc.CreateOrder(context.Background(), "user-1", items)
		s.Require().NoError(err)
		s.Require().Same(wantOrder, order)
	})

	s.Run(testCaseSuccess_CreateOrder_CallOrder, func() {
		d := s.newDeps()
		items := []model.BookingItem{{TicketTypeID: "tt-1", Quantity: 1}}
		var callOrder []string

		d.locker.On("AcquireMany", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) { callOrder = append(callOrder, "AcquireMany") }).
			Return(lock.Unlock(noopUnlock), nil)
		d.repo.On("GetEventIDsForTicketTypes", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) { callOrder = append(callOrder, "GetEventIDsForTicketTypes") }).
			Return(map[string]string{"tt-1": "ev-1"}, nil)
		d.eventCl.On("GetEvent", mock.Anything, "ev-1").
			Run(func(args mock.Arguments) { callOrder = append(callOrder, "GetEvent") }).
			Return(&eventclient.Event{ID: "ev-1", Status: "published"}, nil)
		d.repo.On("CreateOrder", mock.Anything, "user-1", items).
			Run(func(args mock.Arguments) { callOrder = append(callOrder, "CreateOrder") }).
			Return(&model.Order{ID: "order-1"}, nil)

		_, err := d.svc.CreateOrder(context.Background(), "user-1", items)
		s.Require().NoError(err)
		s.Assert().Equal([]string{"AcquireMany", "GetEventIDsForTicketTypes", "GetEvent", "CreateOrder"}, callOrder)
	})
}

func (s *BookingServiceTestSuite) TestConfirmPayment() {
	s.Run(testCaseSuccess_ConfirmPayment_Passthrough, func() {
		d := s.newDeps()
		want := &model.Order{ID: "order-1", Status: model.OrderStatusPaid}
		d.repo.On("ConfirmPayment", mock.Anything, "order-1").Return(want, nil)

		got, err := d.svc.ConfirmPayment(context.Background(), "order-1")
		s.Require().NoError(err)
		s.Require().Same(want, got)
	})

	s.Run(testCaseError_ConfirmPayment_RepoError, func() {
		d := s.newDeps()
		wantErr := apperr.WithMessage(apperr.ErrConflict, "Đơn hàng không ở trạng thái chờ thanh toán")
		d.repo.On("ConfirmPayment", mock.Anything, "order-2").Return(nil, wantErr)

		got, err := d.svc.ConfirmPayment(context.Background(), "order-2")
		s.Require().Nil(got)
		s.Require().Same(wantErr, err)
	})
}

func (s *BookingServiceTestSuite) TestFailPayment() {
	s.Run(testCaseSuccess_FailPayment_Passthrough, func() {
		d := s.newDeps()
		want := &model.Order{ID: "order-1", Status: model.OrderStatusCancelled}
		d.repo.On("FailPayment", mock.Anything, "order-1").Return(want, nil)

		got, err := d.svc.FailPayment(context.Background(), "order-1")
		s.Require().NoError(err)
		s.Require().Same(want, got)
	})

	s.Run(testCaseError_FailPayment_RepoError, func() {
		d := s.newDeps()
		wantErr := apperr.ErrNotFound
		d.repo.On("FailPayment", mock.Anything, "order-missing").Return(nil, wantErr)

		got, err := d.svc.FailPayment(context.Background(), "order-missing")
		s.Require().Nil(got)
		s.Require().Same(wantErr, err)
	})
}

func (s *BookingServiceTestSuite) TestGetOwnerID() {
	s.Run(testCaseSuccess_GetOwnerID_ReturnsUserID, func() {
		d := s.newDeps()
		d.repo.On("GetOrderByID", mock.Anything, "order-1").
			Return(&model.Order{ID: "order-1", UserID: "user-42"}, nil)

		owner, err := d.svc.GetOwnerID(context.Background(), "order-1")
		s.Require().NoError(err)
		s.Assert().Equal("user-42", owner)
	})

	s.Run(testCaseError_GetOwnerID_RepoError, func() {
		d := s.newDeps()
		d.repo.On("GetOrderByID", mock.Anything, "order-missing").Return(nil, apperr.ErrNotFound)

		owner, err := d.svc.GetOwnerID(context.Background(), "order-missing")
		s.Require().Error(err)
		s.Assert().Empty(owner)
	})
}

func (s *BookingServiceTestSuite) TestGetOrderByID() {
	s.Run(testCaseSuccess_GetOrderByID_Passthrough, func() {
		d := s.newDeps()
		want := &model.Order{ID: "order-1"}
		d.repo.On("GetOrderByID", mock.Anything, "order-1").Return(want, nil)

		got, err := d.svc.GetOrderByID(context.Background(), "order-1")
		s.Require().NoError(err)
		s.Require().Same(want, got)
	})
}

func (s *BookingServiceTestSuite) TestListMyOrders() {
	s.Run(testCaseSuccess_ListMyOrders_PassesLimitOffset, func() {
		d := s.newDeps()
		p := pagination.Params{Page: 3, PageSize: 10} // -> offset 20, limit 10
		wantOrders := []*model.Order{{ID: "o1"}, {ID: "o2"}}
		d.repo.On("ListOrdersByUser", mock.Anything, "user-1", "pending", 10, 20).
			Return(wantOrders, int64(2), nil)

		orders, total, err := d.svc.ListMyOrders(context.Background(), "user-1", "pending", p)
		s.Require().NoError(err)
		s.Assert().Equal(wantOrders, orders)
		s.Assert().Equal(int64(2), total)
	})

	s.Run(testCaseError_ListMyOrders_RepoError, func() {
		d := s.newDeps()
		p := pagination.Params{Page: 1, PageSize: 20}
		repoErr := errors.New("db down")
		d.repo.On("ListOrdersByUser", mock.Anything, "user-2", "", 20, 0).
			Return(nil, int64(0), repoErr)

		orders, total, err := d.svc.ListMyOrders(context.Background(), "user-2", "", p)
		s.Require().Same(repoErr, err)
		s.Assert().Nil(orders)
		s.Assert().Zero(total)
	})
}

// isAppErrCode asserts err is an *apperr.Error with the given Code — the
// convention's sanctioned way to check business errors (see
// backend-conventions.md: *apperr.Error has no Is(target) method yet, so
// errors.Is(err, apperr.ErrX) does not work here).
func isAppErrCode(err error, code string) bool {
	var appErr *apperr.Error
	return errors.As(err, &appErr) && appErr.Code == code
}
