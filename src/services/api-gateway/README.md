# API Gateway

**Trạng thái:** chưa implement — placeholder scaffold.

Routing REST → gRPC tới các service phía sau, xác thực JWT, rate limit, CORS. Không sở hữu database nghiệp vụ — chỉ dùng Redis cho rate-limit state.

- Kiến trúc & vai trò chi tiết: [../../../docs/01-architecture/system-architecture.md](../../../docs/01-architecture/system-architecture.md)
- Auth & JWT: [../../../docs/04-security/authentication.md](../../../docs/04-security/authentication.md)
- Rate limiting: [../../../docs/04-security/rate-limiting.md](../../../docs/04-security/rate-limiting.md)

Khi implement, tuân theo layout `cmd/`, `internal/{handler,service,repository,model}/` mô tả tại [../../../docs/01-architecture/repo-structure.md](../../../docs/01-architecture/repo-structure.md).
