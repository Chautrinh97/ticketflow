# Domain: Payment

**Service sở hữu:** Payment Service · **Database:** PostgreSQL (`payments`) · **Phase:** Mock ở Phase 1 (MVP), tích hợp cổng thanh toán thật ở Phase 2.

## Phạm vi & trách nhiệm

Khởi tạo phiên thanh toán cho một đơn hàng, tích hợp với cổng thanh toán bên thứ ba (Stripe/VNPay/Momo...), và xử lý webhook callback xác nhận kết quả thanh toán một cách an toàn (xác thực chữ ký) và idempotent (không xử lý trùng một giao dịch).

## Data model

Bảng `payments` — xem [../../03-data/postgres-schema.md](../../03-data/postgres-schema.md#payment-service--payments).

## Luồng nghiệp vụ

### Khởi tạo thanh toán

1. Sau khi Booking Service tạo `order` (`status='pending'`), client gọi `POST /payments/:orderId/checkout`.
2. Payment Service tạo bản ghi `payments` (`status='initiated'`), gọi API của cổng thanh toán để tạo phiên thanh toán, trả về URL/thông tin thanh toán cho client redirect tới.

### Webhook callback

1. Cổng thanh toán gọi `POST /payments/webhook` sau khi người dùng hoàn tất (hoặc huỷ) thanh toán.
2. **Xác thực chữ ký HMAC** trước tiên — từ chối (không xử lý payload) nếu chữ ký sai. Chi tiết nguyên tắc xem [../../04-security/rate-limiting.md](../../04-security/rate-limiting.md#webhook-thanh-toán--xác-thực--idempotency).
3. **Kiểm tra idempotency theo `provider_txn_id`**: nếu giao dịch này đã được xử lý xong trước đó (status cuối cùng đã ghi nhận), trả `200 OK` ngay, không xử lý lại.
4. Cập nhật `payments.status` (`success`/`failed`), lưu `raw_payload` (nguyên văn payload từ gateway, phục vụ audit/debug khi có tranh chấp).
5. Phát `payment.success` hoặc `payment.failed` cho Booking Service (đồng bộ ở Phase 1-2, qua message queue ở Phase 4) và Notification Service.

### Mock ở Phase 1

Ở MVP, có thể thay bước gọi cổng thanh toán thật bằng một endpoint nội bộ tự đánh dấu `payments.status='success'` ngay lập tức (giả lập thanh toán thành công tức thời) — mục tiêu là luồng `order.created → payment.success → ticket.issued` chạy đúng đắn về mặt state machine, chưa cần tích hợp gateway thật. Khi lên Phase 2, thay thế phần gọi gateway thật mà **không đổi** hợp đồng API/webhook đã định nghĩa.

## Idempotency — vì sao bắt buộc

Cổng thanh toán có thể gọi lại cùng một webhook nhiều lần (retry do timeout, do mạng...). Nếu xử lý không idempotent, hệ thống có thể phát `ticket.issued` hai lần cho cùng một đơn hàng (gửi email trùng, hoặc tệ hơn là sinh trùng bản ghi `tickets` nếu logic sinh vé không kiểm tra trạng thái `orders.status` hiện tại trước khi sinh).

## Quan hệ với domain khác

- **booking**: nhận `order.created` (gián tiếp qua việc booking service gọi checkout), publish `payment.success`/`payment.failed` để Booking Service chuyển trạng thái đơn hàng và sinh vé.
- **notification**: publish các event trên để Notification Service gửi email xác nhận/thông báo thất bại.

## API liên quan

Xem [../../../api-docs/openapi/payment-service.yaml](../../../api-docs/openapi/payment-service.yaml).

## Phân theo phase

| Tính năng | Phase |
|---|---|
| Mock thanh toán, state machine `initiated → success/failed` | 1 (MVP) |
| Tích hợp cổng thanh toán thật, webhook xác thực chữ ký + idempotent | 2 (Must-have) |
| Publish qua message queue thay vì gọi trực tiếp | 4 (Nice-to-have) |
