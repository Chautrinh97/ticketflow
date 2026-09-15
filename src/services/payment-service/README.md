# Payment Service

**Trạng thái:** Phase 1 (MVP) — đã implement: `POST /payments/:orderId/checkout` (mock — xác nhận thành công ngay trong request, gọi gRPC `ConfirmOrderPayment` sang Booking Service), `POST /payments/webhook` (pass-through mỏng, chưa xác thực chữ ký — dùng để test thủ công luồng thất bại). Không có gRPC server riêng (không service nào gọi payment-service qua gRPC ở Phase 1). Chưa implement: tích hợp cổng thanh toán thật, xác thực chữ ký HMAC + idempotent theo `provider_txn_id` (Phase 2).

Tích hợp cổng thanh toán, xử lý webhook. Database: PostgreSQL (`payments`).

- Domain spec: [../../../docs/02-domains/payment/spec.md](../../../docs/02-domains/payment/spec.md)
- Hợp đồng API: [../../../api-docs/openapi/payment-service.yaml](../../../api-docs/openapi/payment-service.yaml)
- Schema dữ liệu: [../../../docs/03-data/postgres-schema.md](../../../docs/03-data/postgres-schema.md)
- Webhook security: [../../../docs/04-security/rate-limiting.md](../../../docs/04-security/rate-limiting.md)

Khi implement, tuân theo layout `cmd/`, `internal/{handler,service,repository,model}/` mô tả tại [../../../docs/01-architecture/repo-structure.md](../../../docs/01-architecture/repo-structure.md).
