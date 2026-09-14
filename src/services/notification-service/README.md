# Notification Service

**Trạng thái:** chưa implement — placeholder scaffold.

Gửi email/push, lưu thông báo in-app. Database: PostgreSQL (`notifications`). Là consumer thuần tuý của event nội bộ — không service nào khác gọi trực tiếp API "gửi thông báo" của service này.

- Domain spec: [../../../docs/02-domains/notification/spec.md](../../../docs/02-domains/notification/spec.md)
- Hợp đồng API: [../../../api-docs/openapi/notification-service.yaml](../../../api-docs/openapi/notification-service.yaml)
- Event-driven design: [../../../docs/01-architecture/event-driven-design.md](../../../docs/01-architecture/event-driven-design.md)

Khi implement, tuân theo layout `cmd/`, `internal/{handler,service,repository,model}/` mô tả tại [../../../docs/01-architecture/repo-structure.md](../../../docs/01-architecture/repo-structure.md).
