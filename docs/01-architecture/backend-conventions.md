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

### Công cụ chuẩn

Thêm vào `go.mod` bằng `go get` khi **lần đầu thực sự cần** — không thêm trước khi có test nào dùng tới (đúng tinh thần "không thêm code/dependency phòng hờ" ở AGENTS.md):

- `github.com/stretchr/testify` (`assert`/`require`/`suite`/`mock`) — assertion + cấu trúc test chuẩn cho mọi test từ nay.
- `github.com/testcontainers/testcontainers-go` (+ module `.../modules/postgres`) — tự spin container Postgres thật cho integration test, thay vì giả định docker-compose Postgres đã chạy sẵn.
- `github.com/golang-migrate/migrate/v4` (+ driver `database/postgres`, source `source/file`) — chạy đúng file `migrations/*.up.sql` thật của service lên container test, thay vì hand-copy schema trong code test.
- `github.com/vektra/mockery/v2` — dev-tool CLI generate mock từ interface (không phải runtime dependency của service).
- `github.com/go-redis/redismock/v9` — mock client `go-redis` (khớp bản `go-redis/v9` đã có sẵn) để test logic dùng Redis mà không cần Redis thật.

### Cấu trúc test

- 1 file test cho 1 file nguồn, cùng package (white-box) — vd `booking_repository.go` ↔ `booking_repository_test.go`.
- 1 `testify.Suite` (`<X>TestSuite`) cho mỗi file test; mỗi hàm/method public cần test có 1 test method chạy suite đó.
- Mỗi tình huống là 1 sub-test qua `s.Run(caseName, func() {...})`; `caseName` khai báo `const` ở đầu file, tiền tố `[Success]`/`[Error]` mô tả rõ kỳ vọng (vd `testCaseError_CreateOrder_InsufficientStock = "[Error] Từ chối khi không đủ tồn kho"`) — tên hằng đóng vai trò tài liệu sống.
- Assertion qua `s.Require()`/`s.Assert()` (testify suite-embedded), không so sánh tay + `t.Fatalf`.
- Không dùng `t.Parallel()` cho test dùng chung 1 container/DB trong cùng package — tránh nhiễu giữa các sub-test không phải là race đang được kiểm thử.

### Integration test (cần Postgres thật)

Chuẩn mới cho **test viết mới** — test hiện có (`booking_repository_concurrency_test.go`) giữ nguyên như hiện tại, không bắt buộc migrate ngược sang cách này:

1. `TestMain(m *testing.M)` cấp package: dùng `testcontainers-go` khởi 1 container Postgres **1 lần cho cả package**, lưu connection string/pool vào biến package-level dùng chung cho mọi suite trong file đó.
2. Áp schema bằng cách chạy thật `migrations/*.up.sql` của chính service qua `golang-migrate/migrate/v4` (driver `postgres` trỏ connection string container + `source/file` trỏ thư mục `migrations/`) — không hand-copy `CREATE TABLE` trong code test nữa, tránh lệch với migration thật theo thời gian.
3. Container không khởi được (không có Docker daemon khả dụng) → log cảnh báo và `os.Exit(0)` ngay trong `TestMain`, coi như skip toàn bộ package — giữ đúng tinh thần "graceful skip" của `t.Skipf` cũ, để môi trường/CI thiếu Docker không bị fail cứng.
4. Dọn container (`container.Terminate(ctx)`) sau `m.Run()`.
5. Seed dữ liệu ban đầu qua helper viết tay (`newTestX`/`seedX`, mẫu `seedTicketType` hiện có) — không cần fixture/golden file ở quy mô repo hiện tại.
6. Giữa các sub-test dùng chung DB: dọn bảng liên quan (`TRUNCATE ... CASCADE`) trong `TearDownTest`/`TearDownSubTest` của suite, để case sau không bị ảnh hưởng bởi case trước.
7. Test race-condition (N goroutine + channel `start` + `sync/atomic` + assert invariant, đúng mẫu `booking_repository_concurrency_test.go`) viết y hệt cách hiện tại — chỉ đổi nguồn lấy pool (container thay vì `TEST_DATABASE_URL`).

### Unit test có mock (service phụ thuộc gRPC client/Redis)

Áp dụng khi muốn test logic ở `service` tách biệt khỏi DB/gRPC/Redis thật.

**Điều kiện cần trước khi mock được**: dependency muốn mock phải là **interface**, không phải struct cụ thể. Hiện tại constructor các `service` (vd `NewBookingService(repo *repository.BookingRepository, locker *lock.TicketTypeLocker, eventCl *eventclient.Client)`) nhận thẳng struct cụ thể — mockery chỉ generate được mock từ interface. Muốn unit-test 1 service với dependency nào đó, **trước tiên** định nghĩa 1 interface hẹp ngay tại package `service` (nơi tiêu thụ, đúng idiom Go "accept interfaces, return structs"), chỉ khai đúng method thực sự dùng tới (vd `type EventChecker interface { GetEvent(ctx context.Context, eventID string) (*eventclient.Event, error) }`), rồi đổi tham số constructor sang nhận interface đó — struct thật (`*eventclient.Client`, `*lock.TicketTypeLocker`) tự động implement, không cần sửa gì ở struct. Đây là thay đổi nhỏ ở code sản xuất, làm khi thực sự bắt đầu unit-test service đó, không làm hàng loạt trước khi cần.

1. Cấu hình `mockery` (file `.mockery.yaml` ở `src/`, tạo khi lần đầu cần) trỏ tới interface vừa định nghĩa, output mock vào subpackage `mocks` cạnh nơi định nghĩa interface (vd `internal/service/mocks`).
2. Dùng mock trong test: `mockEventCl := mocks.NewEventChecker(t)`; `.On("GetEvent", mock.Anything, eventID).Return(&eventclient.Event{...}, nil)`.
3. Với `internal/lock` (bọc redsync qua `go-redis`): dùng `redismock.NewClientMock()` tạo `*redis.Client` giả, truyền thẳng vào `lock.NewTicketTypeLocker(client)` (constructor đã nhận đúng `*redis.Client`, không cần đổi gì vì `redismock` trả về đúng type thật, chỉ intercept ở tầng transport) — cuối test gọi `redisMock.ExpectationsWereMet()`.
4. Assert lỗi nghiệp vụ: tiếp tục dùng `errors.As` + so `.Code` với biến `apperr` có sẵn (như test hiện tại) — **không** dùng `s.Require().ErrorIs(err, apperr.ErrConflict)`, vì `*apperr.Error` là con trỏ được `New`/`WithMessage` tạo mới mỗi lần, chưa có method `Is(target error) bool` so theo `Code` để `errors.Is` nhận diện đúng. Muốn dùng đúng idiom đó cần thêm method `Is` cho `apperr.Error` trước — việc riêng, ngoài phạm vi quy ước này.

### Quy tắc chung

- Với mọi luồng nghiệp vụ có tính transaction đụng tồn kho/số dư (đặt vé, thanh toán — nêu đích danh trong `AGENTS.md`): **bắt buộc** có test race-condition/concurrent-request trước khi coi là hoàn tất, không chỉ test happy-path (`AGENTS.md` mục "Kiểm tra trước khi coi một thay đổi là hoàn tất").
- Test logic thuần (validation, mapping lỗi sang `apperr`, tính toán) không chạm DB/gRPC/Redis: `testify/assert` bình thường, không cần suite/mock, chạy được ngay bằng `go test ./...` không cần Docker hay biến môi trường nào.

## Khi nào thêm code mới vào `src/pkg/`

Chỉ khi có **từ 2 service trở lên đang thực sự cần dùng ngay** — không thêm trước "phòng khi cần" (đúng tinh thần `AGENTS.md` quy ước #5 áp dụng cho cả code hạ tầng, không riêng business logic). Code chỉ 1 domain cần thì để trong `internal/` của chính service đó.
