---
name: go-test
description: Write Go tests for a backend service — integration tests against a real containerized Postgres, unit tests with testify/mockery/redismock mocks, or plain unit tests — including the mandatory race-condition/concurrency test for transactional flows (booking, payment) required by AGENTS.md's completion checklist. Use when adding tests, or when a transactional flow has no concurrency coverage yet.
---

# go-test

Viết test Go cho service dưới `src/services/<name>/`. **Luôn đọc mục "Testing" trong [docs/01-architecture/backend-conventions.md](../../../docs/01-architecture/backend-conventions.md) trước** — đó là nơi chứa toàn bộ quy ước chi tiết (công cụ chuẩn, cấu trúc test, mẫu `TestMain`/testcontainers, cách mock từng loại dependency). File skill này chỉ mô tả quy trình áp dụng, không lặp lại cơ chế.

## Quy trình

1. **Xác định loại test cần viết**:
   - **Integration** (repository, hoặc luồng nghiệp vụ xuyên layer cần Postgres thật) → dùng đúng mẫu `TestMain` + testcontainers-go + golang-migrate ở `backend-conventions.md`.
   - **Unit có mock** (`service` phụ thuộc gRPC client/Redis mà muốn cô lập khỏi dependency thật) → kiểm tra dependency cần mock đã là **interface** chưa. Chưa có thì định nghĩa interface hẹp tại package `service` + đổi constructor sang nhận interface đó **trước**, rồi mới generate mock bằng mockery (hoặc dùng `redismock` cho Redis, không cần đổi gì vì constructor `lock.NewTicketTypeLocker` đã nhận đúng `*redis.Client`).
   - **Unit thuần** (validation, mapping lỗi, tính toán không chạm DB/gRPC/Redis) → `testify/assert` bình thường, chạy ngay bằng `go test ./...`, không cần Docker/biến môi trường.
2. Dependency chưa có trong `go.mod` (testify/testcontainers-go/golang-migrate/mockery/redismock) → thêm bằng `go get` ngay trong lần viết test này — không thêm trước khi thực sự cần.
3. Mọi luồng nghiệp vụ có tính transaction đụng tồn kho/số dư (đặt vé, thanh toán — nêu đích danh trong AGENTS.md) **phải** có test race-condition/concurrent-request, không chỉ happy-path, trước khi coi thay đổi là hoàn tất (AGENTS.md mục "Kiểm tra trước khi coi một thay đổi là hoàn tất"). Điều này bao gồm cả `ConfirmPayment`/`FailPayment` trong `booking-service` (không chỉ `CreateBooking`) — gọi đồng thời confirm/fail/huỷ trên cùng order không được gây lệch trạng thái hoặc lệch `sold_count`.
4. **Phase 2 — job/webhook idempotency**: cron job (huỷ đơn quá hạn, snapshot doanh thu ngày) và webhook handler (payment) phải có test chạy **hai lần liên tiếp trên cùng input** (cùng `order_id` đã expired, cùng payload webhook) và assert kết quả lần 2 không làm thay đổi thêm state (không trừ/cộng `sold_count` hay ghi nhận thanh toán 2 lần) — đặt cạnh mandate race-condition ở mục 3, cùng nguyên tắc "không chỉ happy-path".
5. Sau khi thêm test, cập nhật dòng trạng thái test trong `README.md` của chính service đó (mirror cách `booking-service/README.md` tự báo cáo trạng thái).

## Lưu ý phạm vi

`booking-service/internal/repository/booking_repository_concurrency_test.go` hiện có **không** theo convention mới (dùng `TEST_DATABASE_URL` thủ công + hand-copy schema) — đây là test hợp lệ, không cần migrate lại; convention mới ở `backend-conventions.md` chỉ áp dụng cho test viết mới từ nay.
