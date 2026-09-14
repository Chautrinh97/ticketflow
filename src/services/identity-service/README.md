# Identity Service

**Trạng thái:** chưa implement — placeholder scaffold.

Đăng ký/đăng nhập, phát JWT, quản lý user & role. Database: PostgreSQL (`users`).

- Domain spec: [../../../docs/02-domains/identity/spec.md](../../../docs/02-domains/identity/spec.md)
- Hợp đồng API: [../../../api-docs/openapi/identity-service.yaml](../../../api-docs/openapi/identity-service.yaml)
- Schema dữ liệu: [../../../docs/03-data/postgres-schema.md](../../../docs/03-data/postgres-schema.md)

Khi implement, tuân theo layout `cmd/`, `internal/{handler,service,repository,model}/` mô tả tại [../../../docs/01-architecture/repo-structure.md](../../../docs/01-architecture/repo-structure.md).
