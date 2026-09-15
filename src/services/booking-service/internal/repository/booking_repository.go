// Package repository implements booking-service's ACID transactional core
// directly on pgx (no ORM — mandated by the SELECT ... FOR UPDATE control
// needed here, see docs/02-domains/booking/spec.md and the implementation
// plan's decision to use raw SQL only in this service).
package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ticketflow/pkg/apperr"
	"ticketflow/services/booking-service/internal/model"
)

const bookingHoldDuration = 15 * time.Minute

type BookingRepository struct {
	pool *pgxpool.Pool
}

func NewBookingRepository(pool *pgxpool.Pool) *BookingRepository {
	return &BookingRepository{pool: pool}
}

// GetEventIDsForTicketTypes is an unlocked read used by the service layer
// to resolve which event(s) own the requested ticket types, before
// acquiring any lock, so it can precheck via event-service's gRPC GetEvent
// that each owning event is 'published'.
func (r *BookingRepository) GetEventIDsForTicketTypes(ctx context.Context, ticketTypeIDs []string) (map[string]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, event_id FROM ticket_types WHERE id = ANY($1)`, ticketTypeIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string, len(ticketTypeIDs))
	for rows.Next() {
		var id, eventID string
		if err := rows.Scan(&id, &eventID); err != nil {
			return nil, err
		}
		result[id] = eventID
	}
	return result, rows.Err()
}

type ticketTypeRow struct {
	id    string
	price float64
	quota int
	sold  int
}

// CreateOrder is the ACID core described in docs/02-domains/booking/spec.md,
// generalized to accept multiple ticket types in one order (per the
// implementation plan's resolved checkout-cart-scope decision): it locks
// every distinct ticket_type_id row (sorted by id — deadlock-safe against
// concurrent multi-item orders touching overlapping types), checks each
// type's remaining stock, and either commits one order covering all items
// or rolls back the whole reservation atomically (all-or-nothing).
func (r *BookingRepository) CreateOrder(ctx context.Context, userID string, items []model.BookingItem) (*model.Order, error) {
	if len(items) == 0 {
		return nil, apperr.WithMessage(apperr.ErrValidation, "Đơn hàng cần ít nhất 1 loại vé")
	}

	qtyByID := make(map[string]int, len(items))
	for _, it := range items {
		if it.Quantity < 1 {
			return nil, apperr.WithMessage(apperr.ErrValidation, "Số lượng vé phải lớn hơn 0")
		}
		qtyByID[it.TicketTypeID] += it.Quantity
	}
	ids := make([]string, 0, len(qtyByID))
	for id := range qtyByID {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Lock rows one at a time, in ascending id order (ids is sorted above),
	// rather than a single `WHERE id = ANY($1) ORDER BY id FOR UPDATE`
	// query: FOR UPDATE locks are acquired as the query's underlying scan
	// visits rows, which is not guaranteed to follow the ORDER BY clause
	// (that reorders the output, not necessarily the lock-acquisition
	// sequence during the scan). Issuing one statement per id in program
	// order is what actually guarantees every concurrent transaction
	// touching an overlapping set of ticket_types acquires locks in the
	// same ascending order — the real deadlock-safety mechanism.
	found := make(map[string]ticketTypeRow, len(ids))
	for _, id := range ids {
		var row ticketTypeRow
		row.id = id
		err := tx.QueryRow(ctx,
			`SELECT price, quota, sold_count FROM ticket_types WHERE id = $1 FOR UPDATE`, id,
		).Scan(&row.price, &row.quota, &row.sold)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, apperr.WithMessage(apperr.ErrNotFound, "Một số loại vé không tồn tại")
			}
			return nil, err
		}
		found[id] = row
	}

	for id, qty := range qtyByID {
		row := found[id]
		if row.quota-row.sold < qty {
			return nil, apperr.WithMessage(apperr.ErrConflict, "Không đủ tồn kho cho loại vé đã chọn")
		}
	}

	for id, qty := range qtyByID {
		if _, err := tx.Exec(ctx, `UPDATE ticket_types SET sold_count = sold_count + $1 WHERE id = $2`, qty, id); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	expiresAt := now.Add(bookingHoldDuration)
	orderID := uuid.NewString()

	var totalAmount float64
	orderItems := make([]model.OrderItem, 0, len(qtyByID))
	for id, qty := range qtyByID {
		unitPrice := found[id].price
		totalAmount += unitPrice * float64(qty)
		orderItems = append(orderItems, model.OrderItem{
			ID: uuid.NewString(), OrderID: orderID, TicketTypeID: id, Quantity: qty, UnitPrice: unitPrice,
		})
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO orders (id, user_id, status, total_amount, expires_at, created_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		orderID, userID, model.OrderStatusPending, totalAmount, expiresAt, now); err != nil {
		return nil, err
	}
	for _, item := range orderItems {
		if _, err := tx.Exec(ctx,
			`INSERT INTO order_items (id, order_id, ticket_type_id, quantity, unit_price) VALUES ($1,$2,$3,$4,$5)`,
			item.ID, item.OrderID, item.TicketTypeID, item.Quantity, item.UnitPrice); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &model.Order{
		ID: orderID, UserID: userID, Status: model.OrderStatusPending,
		TotalAmount: totalAmount, ExpiresAt: expiresAt, CreatedAt: now, Items: orderItems,
	}, nil
}

func (r *BookingRepository) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	var o model.Order
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, status, total_amount, expires_at, created_at FROM orders WHERE id = $1`, id,
	).Scan(&o.ID, &o.UserID, &o.Status, &o.TotalAmount, &o.ExpiresAt, &o.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrNotFound
		}
		return nil, err
	}
	if err := r.hydrate(ctx, []*model.Order{&o}); err != nil {
		return nil, err
	}
	return &o, nil
}

// ListOrdersByUser paginates an order history; status="" means every status.
func (r *BookingRepository) ListOrdersByUser(ctx context.Context, userID, status string, limit, offset int) ([]*model.Order, int64, error) {
	where := "user_id = $1"
	args := []any{userID}
	if status != "" {
		where += " AND status = $2"
		args = append(args, status)
	}

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM orders WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataSQL := fmt.Sprintf(
		"SELECT id, user_id, status, total_amount, expires_at, created_at FROM orders WHERE %s ORDER BY created_at DESC OFFSET $%d LIMIT $%d",
		where, len(args)+1, len(args)+2)
	rows, err := r.pool.Query(ctx, dataSQL, append(args, offset, limit)...)
	if err != nil {
		return nil, 0, err
	}
	var orders []*model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalAmount, &o.ExpiresAt, &o.CreatedAt); err != nil {
			rows.Close()
			return nil, 0, err
		}
		orders = append(orders, &o)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if err := r.hydrate(ctx, orders); err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

// hydrate fills Items (and, for items with tickets, Tickets) on a batch of
// orders with two extra queries instead of one per order.
func (r *BookingRepository) hydrate(ctx context.Context, orders []*model.Order) error {
	if len(orders) == 0 {
		return nil
	}
	orderIDs := make([]string, len(orders))
	byID := make(map[string]*model.Order, len(orders))
	for i, o := range orders {
		orderIDs[i] = o.ID
		byID[o.ID] = o
	}

	itemRows, err := r.pool.Query(ctx,
		`SELECT id, order_id, ticket_type_id, quantity, unit_price FROM order_items WHERE order_id = ANY($1)`, orderIDs)
	if err != nil {
		return err
	}
	var itemIDs []string
	for itemRows.Next() {
		var it model.OrderItem
		if err := itemRows.Scan(&it.ID, &it.OrderID, &it.TicketTypeID, &it.Quantity, &it.UnitPrice); err != nil {
			itemRows.Close()
			return err
		}
		byID[it.OrderID].Items = append(byID[it.OrderID].Items, it)
		itemIDs = append(itemIDs, it.ID)
	}
	itemRows.Close()
	if err := itemRows.Err(); err != nil {
		return err
	}
	if len(itemIDs) == 0 {
		return nil
	}

	byItemID := make(map[string]*model.Order, len(itemIDs))
	for _, o := range orders {
		for _, it := range o.Items {
			byItemID[it.ID] = o
		}
	}

	ticketRows, err := r.pool.Query(ctx,
		`SELECT id, order_item_id, ticket_code, status, issued_at FROM tickets WHERE order_item_id = ANY($1)`, itemIDs)
	if err != nil {
		return err
	}
	defer ticketRows.Close()
	for ticketRows.Next() {
		var t model.Ticket
		if err := ticketRows.Scan(&t.ID, &t.OrderItemID, &t.TicketCode, &t.Status, &t.IssuedAt); err != nil {
			return err
		}
		if o, ok := byItemID[t.OrderItemID]; ok {
			o.Tickets = append(o.Tickets, t)
		}
	}
	return ticketRows.Err()
}

// ConfirmPayment transitions orders.status pending->paid and generates one
// ticket row per unit of order_items.quantity, inside its own transaction —
// called by payment-service via the booking.proto ConfirmOrderPayment RPC
// after a (mocked, in Phase 1) successful payment.
func (r *BookingRepository) ConfirmPayment(ctx context.Context, orderID string) (*model.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM orders WHERE id = $1 FOR UPDATE`, orderID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrNotFound
		}
		return nil, err
	}
	if status != model.OrderStatusPending {
		return nil, apperr.WithMessage(apperr.ErrConflict, "Đơn hàng không ở trạng thái chờ thanh toán")
	}

	if _, err := tx.Exec(ctx, `UPDATE orders SET status = $1 WHERE id = $2`, model.OrderStatusPaid, orderID); err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, `SELECT id, quantity FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	type lineItem struct {
		id  string
		qty int
	}
	var lines []lineItem
	for rows.Next() {
		var li lineItem
		if err := rows.Scan(&li.id, &li.qty); err != nil {
			rows.Close()
			return nil, err
		}
		lines = append(lines, li)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	now := time.Now()
	var tickets []model.Ticket
	for _, li := range lines {
		for i := 0; i < li.qty; i++ {
			t := model.Ticket{
				ID: uuid.NewString(), OrderItemID: li.id, TicketCode: newTicketCode(),
				Status: model.TicketStatusValid, IssuedAt: now,
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO tickets (id, order_item_id, ticket_code, status, issued_at) VALUES ($1,$2,$3,$4,$5)`,
				t.ID, t.OrderItemID, t.TicketCode, t.Status, t.IssuedAt); err != nil {
				return nil, err
			}
			tickets = append(tickets, t)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &model.Order{ID: orderID, Status: model.OrderStatusPaid, Tickets: tickets}, nil
}

// FailPayment runs the compensating transaction: restores sold_count for
// every reserved ticket_type and transitions orders.status to cancelled.
// The sold_count adjustment is a single atomic relative UPDATE (no prior
// SELECT...FOR UPDATE needed here, unlike CreateOrder, since it doesn't
// branch on the read value — it always restores exactly what this order
// reserved).
func (r *BookingRepository) FailPayment(ctx context.Context, orderID string) (*model.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM orders WHERE id = $1 FOR UPDATE`, orderID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrNotFound
		}
		return nil, err
	}
	if status != model.OrderStatusPending {
		return nil, apperr.WithMessage(apperr.ErrConflict, "Đơn hàng không ở trạng thái chờ thanh toán")
	}

	rows, err := tx.Query(ctx, `SELECT ticket_type_id, quantity FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	type lineItem struct {
		ticketTypeID string
		qty          int
	}
	var lines []lineItem
	for rows.Next() {
		var li lineItem
		if err := rows.Scan(&li.ticketTypeID, &li.qty); err != nil {
			rows.Close()
			return nil, err
		}
		lines = append(lines, li)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, li := range lines {
		if _, err := tx.Exec(ctx, `UPDATE ticket_types SET sold_count = sold_count - $1 WHERE id = $2`, li.qty, li.ticketTypeID); err != nil {
			return nil, err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE orders SET status = $1 WHERE id = $2`, model.OrderStatusCancelled, orderID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &model.Order{ID: orderID, Status: model.OrderStatusCancelled}, nil
}

func newTicketCode() string {
	return strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", "")[:16])
}
