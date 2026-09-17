# Payment Service

**Trạng thái:** Phase 1 (MVP) — đã implement: `POST /payments/:orderId/checkout` (mock — xác nhận thành công ngay trong request, gọi gRPC `ConfirmOrderPayment` sang Booking Service), `POST /payments/webhook` (pass-through mỏng, chưa xác thực chữ ký — dùng để test thủ công luồng thất bại). Không có gRPC server riêng (không service nào gọi payment-service qua gRPC ở Phase 1). Chưa implement: tích hợp cổng thanh toán thật, xác thực chữ ký HMAC + idempotent theo `provider_txn_id` (Phase 2).

**Test:** đầy đủ theo `docs/01-architecture/backend-conventions.md` — unit-với-mock cho `internal/service` (Checkout/markSuccess/markFailed/HandleWebhook, qua interface `PaymentStore`/`BookingClient` + mockery) và `internal/handler/http` (qua interface `PaymentUseCase`); pure-unit cho `internal/config`; integration cho `internal/repository` (testcontainers Postgres + migrations thật của cả identity/event/booking/payment, vì `payments.order_id` FK sang `orders`). Có test race-condition (`TestUpdateStatus_ConcurrentCalls_NoGuardOrIdempotencyCheck`) — theo đúng góc nhìn hiện tại: `PaymentRepository.UpdateStatus` **chưa có status guard/optimistic lock/idempotency check** (unconditional `UPDATE ... WHERE id = ?`), nên mọi request đồng thời đều "thành công", state cuối cùng phụ thuộc thứ tự commit vật lý — đây là gap sản xuất thật, chưa được vá (out of scope của lần thêm test này), chỉ được test hoá để không tái phát hiện âm thầm sau này. Chạy: `go test ./services/payment-service/...`.

Tích hợp cổng thanh toán, xử lý webhook. Database: PostgreSQL (`payments`).

- Domain spec: [../../../docs/02-domains/payment/spec.md](../../../docs/02-domains/payment/spec.md)
- Hợp đồng API: [../../../api-docs/openapi/payment-service.yaml](../../../api-docs/openapi/payment-service.yaml)
- Schema dữ liệu: [../../../docs/03-data/postgres-schema.md](../../../docs/03-data/postgres-schema.md)
- Webhook security: [../../../docs/04-security/rate-limiting.md](../../../docs/04-security/rate-limiting.md)

Khi implement, tuân theo layout `cmd/`, `internal/{handler,service,repository,model}/` mô tả tại [../../../docs/01-architecture/repo-structure.md](../../../docs/01-architecture/repo-structure.md).
