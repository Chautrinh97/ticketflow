# Event Service

**Trạng thái:** chưa implement — placeholder scaffold.

CRUD sự kiện, loại vé, tìm kiếm cơ bản. Database: PostgreSQL (`events`, `ticket_types`) + MongoDB (`event_catalog`).

- Domain spec: [../../../docs/02-domains/event-catalog/spec.md](../../../docs/02-domains/event-catalog/spec.md)
- Hợp đồng API: [../../../api-docs/openapi/event-service.yaml](../../../api-docs/openapi/event-service.yaml)
- Schema dữ liệu: [../../../docs/03-data/postgres-schema.md](../../../docs/03-data/postgres-schema.md), [../../../docs/03-data/mongodb-schema.md](../../../docs/03-data/mongodb-schema.md)

Khi implement, tuân theo layout `cmd/`, `internal/{handler,service,repository,model}/` mô tả tại [../../../docs/01-architecture/repo-structure.md](../../../docs/01-architecture/repo-structure.md).
