# API Gateway

**Trạng thái:** Phase 1 (MVP) — đã implement: reverse proxy REST → REST tới từng service phía sau (giữ nguyên path, mỗi service tự phục vụ đúng OpenAPI của nó), CORS theo 1 origin cấu hình được, và kiểm tra phiên đăng nhập cho request có bearer token (local verify JWT + gRPC `CheckSession` sang Identity Service) — token thiếu/không hợp lệ vẫn được forward để mỗi service tự quyết định route nào cần auth. Chưa implement: rate limit (Redis token bucket) — Phase 2, nên gateway hiện chưa dùng Redis.

Routing REST tới từng service phía sau (mỗi service tự phục vụ đúng OpenAPI của nó); gRPC chỉ dùng cho lời gọi nội bộ cụ thể không thuộc hợp đồng public nào — hiện tại là `CheckSession` sang Identity Service. Xác thực JWT, rate limit (Phase 2), CORS. Không sở hữu database nghiệp vụ.

- Kiến trúc & vai trò chi tiết: [../../../docs/01-architecture/system-architecture.md](../../../docs/01-architecture/system-architecture.md)
- Auth & JWT: [../../../docs/04-security/authentication.md](../../../docs/04-security/authentication.md)
- Rate limiting: [../../../docs/04-security/rate-limiting.md](../../../docs/04-security/rate-limiting.md)

Khi implement, tuân theo layout `cmd/`, `internal/{handler,service,repository,model}/` mô tả tại [../../../docs/01-architecture/repo-structure.md](../../../docs/01-architecture/repo-structure.md).
