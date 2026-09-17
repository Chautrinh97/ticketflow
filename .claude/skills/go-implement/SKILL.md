---
name: go-implement
description: Implement or modify a Go backend handler/service/repository under src/services/<name>/ following repo-structure.md layering and backend-conventions.md's shared-package usage rules. Use when writing new Go endpoints, business logic, or repository code for any TicketFlow backend service.
---

# go-implement

Implement code Go dưới `src/services/<name>/`. **Luôn đọc trước** [docs/01-architecture/repo-structure.md](../../../docs/01-architecture/repo-structure.md) (layout thư mục) và [docs/01-architecture/backend-conventions.md](../../../docs/01-architecture/backend-conventions.md) (cách dùng `src/pkg/`) — skill này không lặp lại nội dung 2 file đó.

## Checklist bắt buộc

1. **OpenAPI-first**: `api-docs/openapi/<service>.yaml` phải đã phản ánh đúng endpoint sắp code (AGENTS.md quy ước #1). Chưa có/chưa khớp → dừng lại, dùng skill `api-spec` cập nhật OpenAPI trước.
2. **Domain spec-first**: đọc `docs/02-domains/<domain>/spec.md` liên quan trước khi viết logic nghiệp vụ — nếu ý định implement khác spec (phát hiện spec thiếu case), dừng lại, cập nhật spec (skill `domain-spec`) hoặc hỏi người dùng, không code trước rồi suy spec sau.
3. **Layering**: `handler` (`http`/`grpc`) không chứa business logic — mọi quyết định nghiệp vụ nằm ở `service`; `repository` là lớp duy nhất chạm DB (ngoại lệ đã biết: `booking-service/internal/repository` dùng raw SQL/pgx thay vì ORM để kiểm soát `SELECT ... FOR UPDATE`, service khác dùng pattern ORM đã có sẵn trong chính service đó).
4. **Lỗi**: `service` trả `*apperr.Error` (dùng biến có sẵn: `ErrNotFound`/`ErrForbidden`/`ErrConflict`/`ErrValidation`/`ErrTooManyRequests`/`ErrUnauthorized`, hoặc `apperr.WithMessage`); `handler/http` gọi `apperr.Respond`; `handler/grpc` bọc bằng `apperr.AsGRPCStatus`/unwrap bằng `apperr.FromGRPCError` — không leak raw error string ra client.
5. **Auth**: gắn `httpauth.RequireAuth` (+ `RequireRole` nếu giới hạn role); khi cần kiểm tra ownership, `resourceLookup` truyền vào `httpauth.RequireOwnership` phải truy vấn tài nguyên thật (DB/gRPC), không tin field client gửi lên — role đúng là chưa đủ (AGENTS.md, `docs/04-security/authorization.md`).
6. **Phân trang**: endpoint list dùng `pagination.ParseParams` + `pagination.New(items, total)` → response `{items,total}` đúng khớp OpenAPI.
7. **gRPC nội bộ**: chỉ dùng cho lời gọi giữa service (không phải hợp đồng public), gắn interceptor logging/recovery 2 chiều (`grpcinterceptor`).
8. **Database-per-service**: không import `internal/model` hay mở kết nối DB của service khác — giao tiếp chéo service qua gRPC nội bộ hoặc message queue theo `docs/01-architecture/system-architecture.md` (AGENTS.md quy ước #3).
9. **Event mới**: nếu luồng publish event type mới, cập nhật `docs/01-architecture/event-driven-design.md` trong cùng thay đổi (AGENTS.md quy ước #4).
10. **Đúng phạm vi phase**: kiểm tra `docs/07-roadmap/phase-N-*.md` — không viết code "phòng hờ" cho nice-to-have chưa tới phase (AGENTS.md quy ước #5).
11. **Migration**: file `migrations/*.up.sql`/`*.down.sql` khớp cột-với-cột với `docs/03-data/postgres-schema.md` (hoặc mongodb-schema.md).
12. Sau khi implement xong luồng transactional (đặt vé, thanh toán, hoặc luồng tương tự đụng tồn kho/số dư): chuyển sang skill `go-test` để viết race-condition test bắt buộc — chưa có test này thì chưa coi là hoàn tất.

## Checklist bổ sung — mối quan tâm Phase 2

Áp dụng khi endpoint/luồng đang implement thuộc các nhóm sau (bỏ qua nếu không liên quan):

13. **Redis cache-aside** (đọc chi tiết/danh sách sự kiện...): đọc [docs/03-data/redis-keys.md](../../../docs/03-data/redis-keys.md) trước để dùng đúng key pattern/TTL đã định nghĩa — không tự đặt key mới. Invalidate cache **ngay trong cùng transaction hoặc ngay sau khi ghi thành công** (không lệch cache), không dựa vào TTL để tự hết hạn thay cho invalidate chủ động khi organizer sửa/xoá resource.
14. **Rate limit**: middleware token bucket chỉ đặt ở API Gateway theo [docs/04-security/rate-limiting.md](../../../docs/04-security/rate-limiting.md) — không tự thêm rate limit rải rác ở service nội bộ (service nội bộ tin tưởng traffic đã qua gateway).
15. **Cron/background job**: tuân theo [docs/05-infra-devops/background-jobs.md](../../../docs/05-infra-devops/background-jobs.md) — chạy trong transaction cùng nguyên tắc `SELECT ... FOR UPDATE` như luồng đặt vé khi đụng tồn kho/số dư, và phải **idempotent** (chạy lại nhiều lần trên cùng bản ghi không gây lệch dữ liệu). Phase 2 chạy bằng scheduler trong-process trên docker-compose, chưa phải Kubernetes CronJob thật (xem file trên).
16. **Presigned URL (file-service)**: tuân theo scope/whitelist/thời hạn hết hạn mô tả trong [docs/02-domains/file-storage/spec.md](../../../docs/02-domains/file-storage/spec.md) — không tạo URL không giới hạn thời gian hoặc không kiểm tra loại file.
17. **Webhook (payment thật)**: bắt buộc xác thực chữ ký (HMAC/signature header) **trước khi** đọc payload, và kiểm tra idempotency theo `provider_txn_id` (đã tồn tại thì bỏ qua, không xử lý lại) theo [docs/02-domains/payment/spec.md](../../../docs/02-domains/payment/spec.md) — không tin bất kỳ webhook nào chưa qua bước xác thực chữ ký.
