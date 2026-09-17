---
name: go-audit
description: Audit existing Go backend code for OpenAPI drift, layering violations, missing ownership checks, missing race-condition tests, cross-service DB access, and leaked raw errors. Use for reviewing backend Go code changes or an entire service, without writing new features.
---

# go-audit

Audit code Go dưới `src/services/<name>/` (hoặc `src/pkg/`). Đọc [docs/01-architecture/repo-structure.md](../../../docs/01-architecture/repo-structure.md) và [docs/01-architecture/backend-conventions.md](../../../docs/01-architecture/backend-conventions.md) trước để biết đúng quy ước cần đối chiếu.

## Giới hạn cần nêu rõ khi báo cáo

Repo **chưa có** `.golangci.yml`/Makefile — skill này không thể chạy linter tự động (không bắt được unused var, `errcheck`, `gosec`, `staticcheck`...). Mọi check dưới đây làm bằng đọc/grep thủ công, tập trung vào rule đặc thù nghiệp vụ của repo mà linter chung không biết — nếu người dùng muốn bắt lỗi Go mức mã nguồn tổng quát, khuyến nghị thêm `golangci-lint` ở một việc riêng.

## Checklist audit

- **Khớp OpenAPI**: với mỗi route trong `internal/handler/http/router.go`, tìm path+method tương ứng trong `api-docs/openapi/<service>.yaml` — flag route có trong code nhưng không có trong YAML (và ngược lại). Chạy cùng skill `api-spec` (chế độ audit) để có coverage đủ 2 chiều.
- **Layering**: flag business logic (điều kiện nghiệp vụ, gọi DB trực tiếp) nằm trong `handler` thay vì `service`; flag `service` tự viết SQL/query thay vì gọi qua `repository` (trừ ngoại lệ `booking-service` đã biết dùng raw SQL ngay trong `repository`).
- **Ownership**: với mọi route dùng `httpauth.RequireOwnership`, kiểm tra hàm `resourceLookup` truyền vào có thực sự truy vấn tài nguyên (DB/gRPC) để lấy `ownerID`, không lấy từ query param/body client tự gửi. Flag route lẽ ra cần ownership check (tài nguyên thuộc về 1 user/organizer cụ thể) nhưng chỉ có `RequireRole` mà thiếu `RequireOwnership`.
- **Lỗi**: flag `c.JSON(...)` trả lỗi trực tiếp thay vì qua `apperr.Respond`/`apperr.JSON`; flag error message lộ chi tiết nội bộ (SQL, stack trace) ra response.
- **Database-per-service**: flag import `internal/model` hoặc DSN của service khác; flag bất kỳ query nào chạm bảng không do service đó sở hữu ngoài carve-out đã ghi rõ trong domain spec (vd booking/`sold_count`).
- **Race-condition test**: với mọi luồng đụng tồn kho/số dư (đặt vé, thanh toán, hoặc tương tự), kiểm tra có file test theo mẫu concurrency (goroutine + `sync/atomic` + assert invariant) hay chưa — chưa có thì flag là vi phạm checklist AGENTS.md, gợi ý chạy skill `go-test`.
- **Locking**: với transaction đụng tồn kho, kiểm tra có `SELECT ... FOR UPDATE` theo đúng thứ tự id tăng dần khi khoá nhiều dòng (tránh deadlock) — flag nếu thiếu `FOR UPDATE`, khoá không theo thứ tự cố định, hoặc thiếu compensating transaction khi thất bại/hết hạn.
- **Event mới**: flag `publish`/emit event type nào trong code mà không có trong `docs/01-architecture/event-driven-design.md`.
- **RBAC endpoint mới (Phase 2)**: với endpoint admin (`/admin/*`) hoặc duyệt organizer, flag nếu thiếu `RequireRole("super_admin")`, hoặc có `RequireRole` nhưng dùng sai role cho phép; các endpoint này không cần `RequireOwnership` (không có khái niệm chủ sở hữu ở phạm vi toàn hệ thống) — flag ngược lại nếu bị áp nhầm ownership check.
- **Cache invalidation**: với endpoint đọc có dùng Redis cache-aside (chi tiết/danh sách sự kiện...), flag nếu endpoint ghi tương ứng (organizer sửa/xoá event, publish...) không invalidate cùng key mà `docs/03-data/redis-keys.md` quy định — dữ liệu cache có thể lệch với DB.
- **Webhook signature**: với handler nhận webhook (payment), flag nếu thiếu bước xác thực chữ ký (HMAC/signature header) trước khi xử lý payload, hoặc xử lý payload trước khi verify; flag nếu thiếu kiểm tra idempotent theo `provider_txn_id` (có thể xử lý trùng cùng 1 webhook được gửi lại).
- **Rate limit bypass**: flag nếu route đặt vé hoặc route nhạy cảm khác được đăng ký cả ở API Gateway lẫn gọi thẳng service nội bộ theo đường không qua gateway (bỏ qua rate limit).
- Báo cáo theo dạng: `[service] <file>:<dòng nếu có> — <vi phạm cụ thể>, tham chiếu quy ước nào`.
