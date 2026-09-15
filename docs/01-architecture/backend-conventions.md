# Quy ước dùng chung cho backend Go

Đọc [repo-structure.md](repo-structure.md) trước cho cấu trúc thư mục (`cmd/internal/{handler,service,repository,model}`) và bảng liệt kê `src/pkg/`. File này mô tả **cách dùng** từng package dùng chung trong `src/pkg/` cho đúng, và quy ước test bắt buộc — không lặp lại nội dung code comment đã có sẵn trong từng package (đọc trực tiếp source khi cần chi tiết implementation).

## Xử lý lỗi (`pkg/apperr`)

- `service` trả lỗi bằng `*apperr.Error` (dùng biến có sẵn `apperr.ErrNotFound`/`ErrForbidden`/`ErrConflict`/`ErrValidation`/`ErrTooManyRequests`/`ErrUnauthorized`/`ErrInternal`, hoặc `apperr.WithMessage(base, "...")` khi cần message cụ thể hơn) — không tự tạo struct lỗi khác.
- `handler/http` nhận lỗi từ `service` và gọi `apperr.Respond(c, err)` (map `*apperr.Error` đúng HTTP status; lỗi khác → `ErrInternal`, không leak message nội bộ ra client) — dùng `apperr.JSON(c, err)` chỉ khi đã chắc `err` là `*apperr.Error`.
- `handler/grpc` bọc lỗi trả về bằng `apperr.AsGRPCStatus(err)`; phía gọi (client nội bộ sang service khác) unwrap lại bằng `apperr.FromGRPCError(err)` để có lại đúng `*apperr.Error` gốc trước khi trả tiếp cho HTTP client của chính nó.
- Shape `{code,message}` phải khớp 1:1 với schema `Error` trong OpenAPI (xem [api-conventions.md](api-conventions.md)) — đây là điểm hai phía (docs & code) phải luôn đồng bộ.

## Auth (`pkg/authclaims` + `pkg/httpauth`)

- Identity Service ký token bằng `authclaims.Sign(secret, userID, role, jti, ttl)`; API Gateway xác thực phiên còn hiệu lực (revocation/tài khoản bị khoá) qua gRPC `CheckSession` sang Identity Service — các service backend khác **tin tưởng gateway đã làm bước này**, chỉ cần verify local bằng `httpauth.RequireAuth(secret)` (kiểm tra chữ ký + hạn token, không gọi lại `CheckSession`).
- `httpauth.RequireRole(roles...)` áp dụng sau `RequireAuth` khi endpoint giới hạn theo role.
- `httpauth.RequireOwnership(resourceLookup)` áp dụng khi role không đủ — `resourceLookup` **bắt buộc phải truy vấn tài nguyên thật** (DB/gRPC) để lấy `ownerID`, không được suy ra ownership từ field client tự gửi lên (vd query param `?userId=`). Đây là cách hiện thực hoá quy tắc ownership trong [../04-security/authorization.md](../04-security/authorization.md) và [../../AGENTS.md](../../AGENTS.md#nguồn-sự-thật). Xem `OwnerLookup` trong `booking-service/internal/handler/http/booking_handler.go` và `event-service/internal/handler/http/organizer_event_handler.go` làm mẫu.
- Đọc claim trong handler qua `httpauth.UserID(c)`/`Role(c)`/`JTI(c)`/`ExpiresAt(c)` — không tự lấy lại từ `gin.Context` bằng key chuỗi thủ công.

## gRPC nội bộ (`pkg/grpcinterceptor`)

- Mọi gRPC server nội bộ (giữa các service, không phải hợp đồng public) gắn `grpcinterceptor.UnaryServerLogging(serviceName)` + `UnaryServerRecovery()`.
- Mọi gRPC client nội bộ gắn `grpcinterceptor.UnaryClientLogging(callerName)`.
- Trace-id được interceptor tự truyền qua metadata giữa các lệnh gọi nội bộ liên tiếp — không tự thêm cơ chế trace-id khác.

## Phân trang (`pkg/pagination`)

- Mọi endpoint trả danh sách dùng `pagination.ParseParams(pageStr, pageSizeStr)` để đọc query `page`/`page_size` (tự clamp về giá trị mặc định/hợp lệ), rồi `repository` dùng `.Offset()`/`.Limit()` cho câu query.
- Trả kết quả bằng `pagination.New(items, total)` → JSON `{items, total}` — **chỉ 2 field này**, khớp đúng response schema `<Resource>ListResponse` phía OpenAPI (xem [api-conventions.md](api-conventions.md)); `page`/`page_size` chỉ là tham số request, không lặp lại trong response.

## Testing

- Với mọi luồng nghiệp vụ có tính transaction đụng tồn kho/số dư (đặt vé, thanh toán — nêu đích danh trong `AGENTS.md`): **bắt buộc** có test race-condition/concurrent-request trước khi coi là hoàn tất, không chỉ test happy-path (`AGENTS.md` mục "Kiểm tra trước khi coi một thay đổi là hoàn tất").
- Viết theo đúng mẫu `booking-service/internal/repository/booking_repository_concurrency_test.go`:
  - Đọc DSN từ `TEST_DATABASE_URL`, mặc định về DSN docker-compose local nếu biến trống.
  - `t.Skipf(...)` (không `t.Fatalf`) khi không kết nối được Postgres — để `go build`/`go vet`/CI không có Postgres vẫn pass.
  - Tự tạo schema tối thiểu cần dùng ngay trong test (`CREATE TABLE IF NOT EXISTS ...`) — không phụ thuộc migration runner đã chạy trước đó.
  - N goroutine cùng chạy 1 hành động qua 1 channel `start` để khởi chạy đồng loạt, đếm kết quả bằng `sync/atomic`, rồi assert invariant nghiệp vụ (vd `sold_count` không bao giờ vượt `quota`) đúng bất kể thứ tự request tới.
- Không thêm thư viện test/mocking mới (`testify`, `gomock`...) khi chưa có trong `go.mod` — dùng `testing` chuẩn + fake tự viết, khớp phong cách test hiện có.

## Khi nào thêm code mới vào `src/pkg/`

Chỉ khi có **từ 2 service trở lên đang thực sự cần dùng ngay** — không thêm trước "phòng khi cần" (đúng tinh thần `AGENTS.md` quy ước #5 áp dụng cho cả code hạ tầng, không riêng business logic). Code chỉ 1 domain cần thì để trong `internal/` của chính service đó.
