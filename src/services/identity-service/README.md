# Identity Service

**Trạng thái:** Phase 1 (MVP) — đã implement: `POST /auth/login|refresh|logout`, `GET/PATCH /users/me`, gRPC `CheckSession` (dùng bởi API Gateway). Firebase mặc định chạy ở chế độ mock (`AUTH_FIREBASE_MODE=mock`) để chạy được mà không cần project Firebase thật. Refresh-token rotation lưu ở Redis (`refresh:{token}`, không có bảng riêng — xem docs/03-data/redis-keys.md). Chưa implement: duyệt organizer, RBAC đầy đủ, audit log (Phase 2).

Đăng ký/đăng nhập, phát JWT, quản lý user & role. Database: PostgreSQL (`users`).

- Domain spec: [../../../docs/02-domains/identity/spec.md](../../../docs/02-domains/identity/spec.md)
- Hợp đồng API: [../../../api-docs/openapi/identity-service.yaml](../../../api-docs/openapi/identity-service.yaml)
- Schema dữ liệu: [../../../docs/03-data/postgres-schema.md](../../../docs/03-data/postgres-schema.md)

Khi implement, tuân theo layout `cmd/`, `internal/{handler,service,repository,model}/` mô tả tại [../../../docs/01-architecture/repo-structure.md](../../../docs/01-architecture/repo-structure.md).
