# Domain: Booking

**Service sở hữu:** Booking Service · **Database:** PostgreSQL (`orders`, `order_items`, `tickets`) · **Phase:** MVP (Phase 1) — đây là domain trung tâm của toàn hệ thống, phải làm đúng ngay từ Phase 1.

## Phạm vi & trách nhiệm

Tạo đơn hàng, đảm bảo **không bán vượt quá `quota`** của một loại vé dù nhiều request đặt vé đến đồng thời (flash-sale), sinh vé (kèm mã QR) sau khi thanh toán thành công, và xử lý huỷ/hoàn tồn kho khi đơn hàng thất bại hoặc hết hạn giữ chỗ.

## Data model

Bảng `orders`, `order_items`, `tickets` — xem [../../03-data/postgres-schema.md](../../03-data/postgres-schema.md#booking-service--orders-order_items-tickets). Booking Service cũng là bên **duy nhất** được `UPDATE ticket_types.sold_count` (bảng này do Event Service sở hữu định nghĩa, nhưng cột `sold_count` là "sổ cái tồn kho" mà Booking Service vận hành trong transaction của chính nó).

## Luồng đặt vé (booking flow) — nơi thể hiện ACID + locking

```
1. Client gửi POST /bookings { ticket_type_id, quantity }
2. API Gateway: xác thực JWT → kiểm tra rate limit (Redis token bucket)
   → nếu vượt hạn mức: trả 429 Too Many Requests
3. Booking Service: acquire Redis distributed lock theo ticket_type_id
   (chặn bớt tải trước khi chạm DB, đặc biệt lúc flash-sale — không thay thế bước 4)
4. Trong 1 PostgreSQL transaction:
     SELECT quota, sold_count FROM ticket_types WHERE id = ? FOR UPDATE;
     -- kiểm tra quota - sold_count >= quantity, nếu không đủ → rollback, trả lỗi
     UPDATE ticket_types SET sold_count = sold_count + quantity WHERE id = ?;
     INSERT INTO orders (status='pending', expires_at = now() + interval '15 minutes');
     INSERT INTO order_items (...);
   COMMIT;
   -- Atomicity + Isolation đảm bảo không bán vượt quota dù nhiều request song song
5. Release Redis lock
6. Publish "order.created" (Phase 1-2: gọi trực tiếp Payment Service; Phase 4: qua message queue)
7. Payment Service tạo phiên thanh toán → trả payment URL cho client
8. Người dùng thanh toán → Payment Gateway gọi webhook
   → Payment Service xác thực chữ ký, cập nhật payments.status
   → phát "payment.success" hoặc "payment.failed"
9. Nếu "payment.success": Booking Service cập nhật orders.status='paid'
   → sinh tickets (mã QR) → phát "ticket.issued"
10. Nếu "payment.failed" HOẶC cron job phát hiện order pending quá hạn:
    compensating transaction → UPDATE ticket_types SET sold_count = sold_count - quantity
    → orders.status='expired'/'cancelled' (hoàn lại vé vào kho)
11. Notification Service nhận "ticket.issued" / "order.cancelled" → gửi email + push + ghi notifications
```

### Vì sao cần cả Redis lock lẫn `SELECT ... FOR UPDATE`

Redis lock là **hàng rào giảm tải ở tầng ứng dụng** — chặn bớt số request chạm tới DB cùng lúc trong kịch bản flash-sale, giảm áp lực lock contention ở Postgres. `SELECT ... FOR UPDATE` trong transaction Postgres mới là cơ chế **quyết định tính đúng đắn cuối cùng** — kể cả nếu Redis lock bị mất (Redis không có persistence, hoặc lock hết hạn quá sớm), transaction Postgres vẫn đảm bảo không có hai transaction cùng đọc-rồi-ghi `sold_count` chồng lên nhau trên cùng một hàng. Không được bỏ `SELECT ... FOR UPDATE` với lý do "đã có Redis lock rồi".

### Huỷ vé bởi người dùng

`POST /bookings/:id/cancel` — chỉ chủ đơn hàng (`order.user_id == current_user.id`) hoặc `super_admin`, và chỉ trong thời hạn cho phép (vd: trước `start_time` của sự kiện tối thiểu N giờ — giá trị N là tham số hệ thống do `super_admin` cấu hình, xem [../../00-overview/project-overview.md](../../00-overview/project-overview.md) mục cấu hình tham số). Khi huỷ: cập nhật `tickets.status='cancelled'`, hoàn `sold_count` tương ứng trong cùng nguyên tắc transaction + lock như trên.

### Cron huỷ đơn quá hạn

Mỗi phút, tìm `orders` có `status='pending'` và `expires_at < now()` → thực hiện đúng compensating transaction ở bước 10 phía trên. Chi tiết vận hành job xem [../../05-infra-devops/background-jobs.md](../../05-infra-devops/background-jobs.md).

## Quan hệ với domain khác

- **event-catalog**: đọc/ghi `ticket_types.quota`/`sold_count`, không sửa các trường khác của `ticket_types`/`events`.
- **payment**: publish `order.created`, nhận `payment.success`/`payment.failed`.
- **notification**: publish `ticket.issued`, `order.cancelled`.
- **identity**: `orders.user_id` tham chiếu `users.id`.

## API liên quan

Xem [../../../api-docs/openapi/booking-service.yaml](../../../api-docs/openapi/booking-service.yaml).

## Phân theo phase

| Tính năng | Phase |
|---|---|
| `POST /bookings` với transaction ACID + lock, `GET /bookings/:id`, lịch sử đặt vé | 1 (MVP) |
| Rate limit riêng cho endpoint đặt vé, cron huỷ đơn quá hạn, huỷ vé bởi user | 2 (Must-have) |
| Event-driven đầy đủ qua message queue thay vì gọi trực tiếp | 4 (Nice-to-have) |
