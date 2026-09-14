# Booking Service

**Trạng thái:** chưa implement — placeholder scaffold.

Tạo đơn, giữ chỗ, đảm bảo ACID khi trừ tồn kho vé. Database: PostgreSQL (`orders`, `order_items`, `tickets`). Đây là domain trung tâm của hệ thống — đọc kỹ spec transaction/locking trước khi implement.

- Domain spec: [../../../docs/02-domains/booking/spec.md](../../../docs/02-domains/booking/spec.md)
- Hợp đồng API: [../../../api-docs/openapi/booking-service.yaml](../../../api-docs/openapi/booking-service.yaml)
- Schema dữ liệu: [../../../docs/03-data/postgres-schema.md](../../../docs/03-data/postgres-schema.md)
- Distributed lock: [../../../docs/03-data/redis-keys.md](../../../docs/03-data/redis-keys.md)

Khi implement, tuân theo layout `cmd/`, `internal/{handler,service,repository,model}/` mô tả tại [../../../docs/01-architecture/repo-structure.md](../../../docs/01-architecture/repo-structure.md). Bắt buộc có test race-condition (nhiều goroutine đặt vé đồng thời trên cùng ticket_type) trước khi coi luồng đặt vé là hoàn tất.
