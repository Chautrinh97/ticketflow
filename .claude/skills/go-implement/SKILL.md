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
