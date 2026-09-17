# API Gateway

**Trạng thái:** Phase 1 (MVP) — đã implement: reverse proxy REST → REST tới từng service phía sau (giữ nguyên path, mỗi service tự phục vụ đúng OpenAPI của nó), CORS theo 1 origin cấu hình được, và kiểm tra phiên đăng nhập cho request có bearer token (local verify JWT + gRPC `CheckSession` sang Identity Service) — token thiếu/không hợp lệ vẫn được forward để mỗi service tự quyết định route nào cần auth. Chưa implement: rate limit (Redis token bucket) — Phase 2, nên gateway hiện chưa dùng Redis.

**Testing:** unit test đầy đủ cho toàn bộ package, chạy được ngay bằng `go test ./services/api-gateway/...` (không cần Docker/DB — gateway không sở hữu database). `internal/middleware/auth_test.go` cô lập `OptionalSessionCheck` khỏi gRPC thật qua interface `SessionChecker` (mock generate bằng mockery ở `internal/middleware/mocks`), phủ đủ 5 nhánh: không có Bearer → pass-through, JWT không hợp lệ/hết hạn → 401, lỗi gọi `CheckSession` → 503, blacklisted/banned → 401, hợp lệ → `Next()`. `internal/middleware/cors_test.go` test thuần bảng cho `CORS` (origin khớp, preflight OPTIONS, origin không khớp). `internal/proxy/reverse_proxy_test.go` test panic khi URL đích không hợp lệ + 1 test proxy thật qua `httptest.Server`. `internal/config/config_test.go` test default/override từng biến môi trường. `internal/router/router_test.go` smoke-test toàn bộ route table trỏ đúng backend. Toàn bộ chạy pass thật (kể cả `-race`), không có test nào skip.

Routing REST tới từng service phía sau (mỗi service tự phục vụ đúng OpenAPI của nó); gRPC chỉ dùng cho lời gọi nội bộ cụ thể không thuộc hợp đồng public nào — hiện tại là `CheckSession` sang Identity Service. Xác thực JWT, rate limit (Phase 2), CORS. Không sở hữu database nghiệp vụ.

- Kiến trúc & vai trò chi tiết: [../../../docs/01-architecture/system-architecture.md](../../../docs/01-architecture/system-architecture.md)
- Auth & JWT: [../../../docs/04-security/authentication.md](../../../docs/04-security/authentication.md)
- Rate limiting: [../../../docs/04-security/rate-limiting.md](../../../docs/04-security/rate-limiting.md)

Khi implement, tuân theo layout `cmd/`, `internal/{handler,service,repository,model}/` mô tả tại [../../../docs/01-architecture/repo-structure.md](../../../docs/01-architecture/repo-structure.md).
