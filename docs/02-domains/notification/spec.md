# Domain: Notification

**Service sở hữu:** Notification Service · **Database:** PostgreSQL (`notifications`) · **Phase:** Phase 2 (Must-have).

## Phạm vi & trách nhiệm

Gửi thông báo cho người dùng qua nhiều kênh (email, in-app, push) khi có sự kiện nghiệp vụ đáng chú ý xảy ra (đặt vé thành công, thanh toán thành công/thất bại, sắp tới giờ diễn ra sự kiện). Notification Service là **consumer thuần tuý** — không service nào khác gọi trực tiếp API của nó để "yêu cầu gửi thông báo"; nó tự lắng nghe event từ các domain khác. Đây là nguyên tắc quan trọng giữ đúng kiến trúc event-driven (xem [../../01-architecture/event-driven-design.md](../../01-architecture/event-driven-design.md)) — kể cả ở Phase 1-2 khi chưa có message queue thật, Notification Service vẫn nên được gọi qua một lớp trung gian (nội bộ, đồng bộ) đóng vai trò "publish event" thay vì các service khác gọi thẳng vào hàm gửi email của nó.

## Data model

Bảng `notifications` — xem [../../03-data/postgres-schema.md](../../03-data/postgres-schema.md#notification-service--notifications).

## Kênh & trigger

| Kênh | Trigger | Công cụ |
|---|---|---|
| Email | `order.created`, `payment.success`, `event.reminder` (T-24h), `order.cancelled` | SendGrid/SES |
| In-app | Tất cả các event trên | Ghi bảng `notifications`, frontend poll hoặc WebSocket |
| Push *(tuỳ chọn)* | `event.reminder` | Firebase Cloud Messaging |

## Luồng xử lý một event

1. Nhận event (vd: `ticket.issued`) kèm `order_id`, `user_id`, và payload liên quan.
2. **Idempotency**: kiểm tra đã xử lý event này chưa (vd: dựa trên `order_id` + loại event) trước khi gửi — event có thể được deliver lại.
3. Ghi bản ghi `notifications` (kênh in-app).
4. Gửi email tương ứng qua provider (SendGrid/SES).
5. *(tuỳ chọn)* Gửi push qua Firebase Cloud Messaging nếu là `event.reminder`.

## Đọc thông báo

- `GET /users/me/notifications` — danh sách thông báo của user hiện tại, có phân trang.
- `PATCH /users/me/notifications/:id/read` — đánh dấu đã đọc, chỉ cho phép trên thông báo thuộc chính user gọi (kiểm tra ownership, xem [../../04-security/authorization.md](../../04-security/authorization.md)).

## Quan hệ với domain khác

Consume event từ **booking** (`order.created`, `ticket.issued`, `order.cancelled`) và **payment** (`payment.success`, `payment.failed`). Không publish event nào cho domain khác.

## API liên quan

Xem [../../../api-docs/openapi/notification-service.yaml](../../../api-docs/openapi/notification-service.yaml).

## Phân theo phase

| Tính năng | Phase |
|---|---|
| Email xác nhận đơn/thanh toán/huỷ vé, in-app notification, `GET/PATCH` endpoint | 2 (Must-have) |
| Nhắc lịch sự kiện (`event.reminder`), push notification | 2-3 |
| Consume qua message queue thật thay vì gọi nội bộ | 4 (Nice-to-have) |
