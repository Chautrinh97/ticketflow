# Identity Service

**Trạng thái:** Phase 1 (MVP) — đã implement: `POST /auth/login|refresh|logout`, `GET/PATCH /users/me`, gRPC `CheckSession` (dùng bởi API Gateway). Firebase mặc định chạy ở chế độ mock (`AUTH_FIREBASE_MODE=mock`) để chạy được mà không cần project Firebase thật. Refresh-token rotation lưu ở Redis (`refresh:{token}`, không có bảng riêng — xem docs/03-data/redis-keys.md). Chưa implement: duyệt organizer, RBAC đầy đủ, audit log (Phase 2).

**Test:** Đầy đủ theo quy ước ở `docs/01-architecture/backend-conventions.md`#Testing — `internal/firebase` (pure, `MockVerifier.Verify`), `internal/config` (pure, default/override), `internal/handler/http` (pure `toUserDTO` + mock-backed handler suites qua interface `AuthService`/`UserService` mới tách ở `interfaces.go`), `internal/service` (mock-backed `AuthServiceTestSuite`/`UserServiceTestSuite` qua interface `UserRepository` mới tách ở `interfaces.go` + mockery mock ở `internal/service/mocks`; `SessionRepository` dùng thật qua `miniredis`, không mock), `internal/repository` (`UserRepositoryTestSuite` — testcontainers Postgres thật + migrate chạy `migrations/*.up.sql`, graceful-skip nếu không có Docker; `SessionRepositoryTestSuite` — `miniredis`, không cần Docker). Chạy `go test ./services/identity-service/... -v`.

Đăng ký/đăng nhập, phát JWT, quản lý user & role. Database: PostgreSQL (`users`).

- Domain spec: [../../../docs/02-domains/identity/spec.md](../../../docs/02-domains/identity/spec.md)
- Hợp đồng API: [../../../api-docs/openapi/identity-service.yaml](../../../api-docs/openapi/identity-service.yaml)
- Schema dữ liệu: [../../../docs/03-data/postgres-schema.md](../../../docs/03-data/postgres-schema.md)

Khi implement, tuân theo layout `cmd/`, `internal/{handler,service,repository,model}/` mô tả tại [../../../docs/01-architecture/repo-structure.md](../../../docs/01-architecture/repo-structure.md).
