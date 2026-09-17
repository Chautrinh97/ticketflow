# Booking Service

**Trạng thái:** Phase 1 (MVP) — đã implement: `POST /bookings` (nhận nhiều dòng `items[]` trong 1 đơn — xem phần "Vì sao `items[]`" bên dưới), `GET /bookings/:id`, `GET /users/me/bookings`, gRPC `GetOrder`/`ConfirmOrderPayment`/`FailOrderPayment` (dùng bởi Payment Service). `internal/repository` dùng raw SQL/pgx (không ORM) để kiểm soát `SELECT ... FOR UPDATE`. Chưa implement: `POST /bookings/:id/cancel`, cron huỷ đơn quá hạn, rate limit riêng, các endpoint buyers/stats (Phase 2).

**Test:** đầy đủ theo quy ước ở `docs/01-architecture/backend-conventions.md` mục "Testing":
- `internal/service`: unit test có mock (`testify/suite` + `mockery`, interface `Repository`/`Locker`/`EventChecker` trích xuất tại package `service`) cho toàn bộ orchestration logic của `CreateOrder`/`ConfirmPayment`/`FailPayment`/`GetOwnerID`/`ListMyOrders` (dedup `ticket_type_id`, thứ tự gọi lock → resolve event → check published → tạo đơn, map lỗi).
- `internal/repository`: `booking_repository_test.go` (MỚI) — integration test chuẩn mới (`TestMain` + `testcontainers-go` Postgres + `golang-migrate` áp thật `migrations/*.up.sql`) cho `GetOrderByID`, `ListOrdersByUser` (phân trang + lọc status), `ConfirmPayment`, `FailPayment` (happy-path + conflict sai trạng thái + hoàn `sold_count`), **kèm 2 test race-condition bắt buộc**: xác nhận thanh toán trùng lặp đồng thời, và Confirm-vs-Fail chạy đua trên cùng 1 đơn — cả hai đã chạy PASS thật với Postgres container. `booking_repository_concurrency_test.go` (đặt vé đồng thời trên cùng ticket_type, không bán vượt quota) giữ nguyên theo cách cũ (`TEST_DATABASE_URL`, tự skip nếu không có Postgres reachable) — không migrate sang testcontainers.
- `internal/lock`: unit test dùng `miniredis` (không phải `redismock`, cần đúng semantics Lua/SET-NX của redsync) — xác nhận `AcquireMany` sắp xếp id trước khi khoá và từ chối ngay (rollback phần đã khoá) khi có tranh chấp, chạy được không cần Docker.
- `internal/handler/http`, `internal/handler/grpc`: unit test có mock (interface `Bookings` trích xuất tại từng package) cho toàn bộ handler/gRPC method; `dto_test.go` test thuần `toOrderDTO`.
- Không có Docker → integration test tự skip (log cảnh báo + thoát 0), không fail cứng CI.

Tạo đơn, giữ chỗ, đảm bảo ACID khi trừ tồn kho vé. Database: PostgreSQL (`orders`, `order_items`, `tickets`). Đây là domain trung tâm của hệ thống — đọc kỹ spec transaction/locking trước khi implement.

### `items[]` — nhiều loại vé trong 1 đơn hàng

Checkout cho phép chọn nhiều loại vé trong 1 đơn (vd: 2 VIP + 3 Standard), khớp UI ở `docs/06-frontend/screens/user/checkout.md`. Transaction khoá tất cả `ticket_types` liên quan bằng các câu lệnh `SELECT ... FOR UPDATE` riêng biệt theo thứ tự `id` tăng dần (không dùng `WHERE id = ANY(...) ORDER BY ... FOR UPDATE` trong 1 câu lệnh — thứ tự khoá thực tế phụ thuộc cách Postgres quét dữ liệu, không phải `ORDER BY` của kết quả trả về) để đảm bảo không xảy ra deadlock giữa các đơn hàng đa loại vé chạm nhau.

- Domain spec: [../../../docs/02-domains/booking/spec.md](../../../docs/02-domains/booking/spec.md)
- Hợp đồng API: [../../../api-docs/openapi/booking-service.yaml](../../../api-docs/openapi/booking-service.yaml)
- Schema dữ liệu: [../../../docs/03-data/postgres-schema.md](../../../docs/03-data/postgres-schema.md)
- Distributed lock: [../../../docs/03-data/redis-keys.md](../../../docs/03-data/redis-keys.md)

Khi implement, tuân theo layout `cmd/`, `internal/{handler,service,repository,model}/` mô tả tại [../../../docs/01-architecture/repo-structure.md](../../../docs/01-architecture/repo-structure.md). Bắt buộc có test race-condition (nhiều goroutine đặt vé đồng thời trên cùng ticket_type) trước khi coi luồng đặt vé là hoàn tất.
