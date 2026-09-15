---
name: go-test
description: Write Go tests for a backend service, including the mandatory race-condition/concurrency test for transactional flows (booking, payment) required by AGENTS.md's completion checklist. Use when adding tests, or when a transactional flow (stock, balance, order state) has no concurrency coverage yet.
---

# go-test

Viết test Go cho service dưới `src/services/<name>/`. Đọc mục "Testing" trong [docs/01-architecture/backend-conventions.md](../../../docs/01-architecture/backend-conventions.md) trước.

## Quy tắc bắt buộc

- Mọi luồng nghiệp vụ có tính transaction đụng tồn kho/số dư (đặt vé, thanh toán — nêu đích danh trong AGENTS.md) **phải** có test race-condition/concurrent-request, không chỉ test happy-path, trước khi coi thay đổi là hoàn tất (AGENTS.md mục "Kiểm tra trước khi coi một thay đổi là hoàn tất").
- **Ưu tiên ngay**: `payment-service` hiện **chưa có test nào** dù được AGENTS.md nêu tên ngang hàng `booking` trong yêu cầu này — đây là vi phạm checklist đang tồn tại sẵn trong code, nếu người dùng yêu cầu chạy skill này mà không chỉ định service cụ thể, hỏi có muốn ưu tiên `payment-service` trước không.

## Mẫu test concurrency (theo đúng `booking-service/internal/repository/booking_repository_concurrency_test.go`)

1. `testPool(t)` helper: đọc DSN từ `TEST_DATABASE_URL`, mặc định về DSN docker-compose local; `t.Skipf(...)` (không `t.Fatalf`) khi không kết nối được Postgres — để `go build`/`go vet`/CI không có Postgres vẫn pass.
2. `setupSchema(t, pool)` helper: tự `CREATE TABLE IF NOT EXISTS` tối thiểu subset schema cần dùng, mirror đúng migration thật — test tự chứa (self-contained), không phụ thuộc migration runner đã chạy trước.
3. Seed dữ liệu ban đầu (helper riêng, vd `seedTicketType`) trả về id cần dùng.
4. N goroutine cùng gọi 1 hành động, đồng bộ khởi chạy bằng 1 channel `start` (`<-start` trong mỗi goroutine, `close(start)` sau khi spawn hết) để tối đa hoá khả năng race thật xảy ra, đếm kết quả bằng `sync/atomic` (thành công/conflict/lỗi khác).
5. Assert: tổng số thành công đúng bằng giới hạn tài nguyên (vd `quota`), tổng conflict đúng bằng phần còn lại, **0 lỗi khác** loại conflict, và đọc lại state cuối cùng trong DB (vd `sold_count`) để xác nhận invariant giữ đúng — không chỉ tin vào giá trị trả về của lời gọi.
6. Dùng `errors.As` + so `Code` với biến `apperr` có sẵn (vd `apperr.ErrConflict.Code`) để phân loại lỗi conflict vs lỗi khác — không so message string.

## Test không cần DB (unit test thường)

- Logic validation, mapping lỗi sang `apperr`, và business rule không chạm DB thật thì viết table-driven test chuẩn `testing`, chạy được bằng `go test ./...` không cần biến môi trường nào.
- Không thêm `testify`/`gomock` hay thư viện test/mocking mới khi `go.mod` chưa có — dùng `testing` chuẩn + fake tự viết, khớp phong cách hiện có trong repo.

## Sau khi thêm test

Cập nhật dòng trạng thái test trong `README.md` của chính service đó (mirror cách `booking-service/README.md` tự báo cáo trạng thái) để người đọc sau không phải tự đi tìm file test.
