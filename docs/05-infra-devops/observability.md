# Observability *(nice-to-have — Phase 4)*

## Mục tiêu

Cho phép trả lời chính xác câu hỏi "bước nào trong một giao dịch xuyên suốt nhiều service (API Gateway → Booking → Payment → Notification) gây chậm hoặc lỗi" — đặc biệt quan trọng vì hệ thống dùng saga choreography (xem [../01-architecture/event-driven-design.md](../01-architecture/event-driven-design.md)), nơi một giao dịch logic trải rộng qua nhiều service độc lập.

## Thành phần

- **Tracing**: nhúng OpenTelemetry SDK vào từng service Go, propagate `trace_id` qua HTTP header / gRPC metadata xuyên suốt toàn bộ chuỗi gọi. Export sang **Grafana Tempo** (hoặc Jaeger).
- **Metrics**: export sang **Prometheus** — latency theo endpoint, tỉ lệ lỗi theo service, số đơn hàng/phút, độ trễ hàng đợi message queue, thời gian giữ distributed lock.
- **Log tập trung**: qua **Grafana Loki**, log có cấu trúc (structured logging, JSON) kèm `trace_id` để nhảy thẳng từ log sang trace tương ứng.

## Dashboard Grafana dự kiến

- Latency p50/p95/p99 theo endpoint, tách riêng endpoint đặt vé (nhạy cảm nhất với tải cao).
- Tỉ lệ lỗi (4xx/5xx) theo service.
- Số đơn hàng tạo mới / thanh toán thành công / huỷ theo thời gian thực.
- Độ trễ tiêu thụ message queue (thời gian từ publish tới consume).

## Kịch bản demo

Dùng trace để chỉ ra chính xác nguyên nhân chậm khi có tải cao trong một lần flash-sale giả lập: DB lock (chờ `SELECT ... FOR UPDATE`), độ trễ gọi payment gateway, hay độ trễ hàng đợi message queue — mỗi nguyên nhân thể hiện bằng một span khác nhau trong cùng một trace.

## Ghi chú triển khai

Đây là hạng mục Phase 4 — không chặn tiến độ các phase trước. Tuy nhiên khi implement service ở Phase 1-2, nên **để sẵn chỗ** cho việc nhúng OpenTelemetry sau này: dùng structured logging ngay từ đầu, không hardcode logic phụ thuộc vào việc thiếu tracing.
