# Event Service

**Trạng thái:** Phase 1 (MVP) — đã implement: `GET /events`, `GET /events/:slug`, `GET/POST /organizer/events`, `GET/PATCH/DELETE /organizer/events/:id` (GET: đọc chi tiết 1 sự kiện theo id, ownership-checked — xem event-catalog/spec.md), `POST /organizer/events/:id/publish`, `POST /organizer/events/:id/ticket-types`, gRPC `GetEvent` (dùng bởi Booking Service). Tạo sự kiện ghi Postgres + Mongo trong cùng luồng, rollback Postgres nếu ghi Mongo thất bại. Chưa implement: `GET /events/search` (full-text/fuzzy), validate `attributes` theo category, cache Redis (Phase 2).

CRUD sự kiện, loại vé, tìm kiếm cơ bản. Database: PostgreSQL (`events`, `ticket_types`) + MongoDB (`event_catalog`).

- Domain spec: [../../../docs/02-domains/event-catalog/spec.md](../../../docs/02-domains/event-catalog/spec.md)
- Hợp đồng API: [../../../api-docs/openapi/event-service.yaml](../../../api-docs/openapi/event-service.yaml)
- Schema dữ liệu: [../../../docs/03-data/postgres-schema.md](../../../docs/03-data/postgres-schema.md), [../../../docs/03-data/mongodb-schema.md](../../../docs/03-data/mongodb-schema.md)

Khi implement, tuân theo layout `cmd/`, `internal/{handler,service,repository,model}/` mô tả tại [../../../docs/01-architecture/repo-structure.md](../../../docs/01-architecture/repo-structure.md).
