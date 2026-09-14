# Payment Service

**Trạng thái:** chưa implement — placeholder scaffold.

Tích hợp cổng thanh toán, xử lý webhook. Database: PostgreSQL (`payments`).

- Domain spec: [../../../docs/02-domains/payment/spec.md](../../../docs/02-domains/payment/spec.md)
- Hợp đồng API: [../../../api-docs/openapi/payment-service.yaml](../../../api-docs/openapi/payment-service.yaml)
- Schema dữ liệu: [../../../docs/03-data/postgres-schema.md](../../../docs/03-data/postgres-schema.md)
- Webhook security: [../../../docs/04-security/rate-limiting.md](../../../docs/04-security/rate-limiting.md)

Khi implement, tuân theo layout `cmd/`, `internal/{handler,service,repository,model}/` mô tả tại [../../../docs/01-architecture/repo-structure.md](../../../docs/01-architecture/repo-structure.md).
