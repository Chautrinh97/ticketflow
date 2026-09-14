# Phase 4 — Nice-to-have

**Ước lượng:** 2–3 tuần. **Mục tiêu:** mở rộng hệ thống với các thành phần thể hiện chiều sâu system design — không bắt buộc để có một portfolio hoàn chỉnh (Phase 2/3 đã đủ), nhưng làm nổi bật khi có thời gian.

## Phạm vi

| Hạng mục | Việc cần làm | Tài liệu liên quan |
|---|---|---|
| Message queue + saga | Thay giao tiếp đồng bộ giữa Booking/Payment/Notification bằng RabbitMQ/Kafka theo mô hình event-driven, saga choreography đầy đủ | [../01-architecture/event-driven-design.md](../01-architecture/event-driven-design.md) |
| Observability | OpenTelemetry tracing xuyên service, Prometheus metrics, Grafana Loki log tập trung, dashboard | [../05-infra-devops/observability.md](../05-infra-devops/observability.md) |
| Elasticsearch | Tách Search Service riêng, đồng bộ từ Postgres/Mongo, fuzzy + autocomplete + xếp hạng theo `_score` | [../02-domains/search/spec.md](../02-domains/search/spec.md) |
| Redis persistence | Cấu hình RDB + AOF cho dữ liệu lock/rate-limit quan trọng lúc cao điểm | [../03-data/redis-keys.md](../03-data/redis-keys.md) |
| Analytics Service | Tách service riêng, ClickHouse cho truy vấn phân tích quy mô lớn | [../02-domains/analytics/spec.md](../02-domains/analytics/spec.md) |
| AI Integration | Chatbot hỗ trợ (RAG), gợi ý sự kiện cá nhân hoá (embedding + pgvector), tự sinh mô tả/tag sự kiện | [../02-domains/ai-integration/spec.md](../02-domains/ai-integration/spec.md) |

## Nguyên tắc

Mỗi hạng mục ở Phase 4 độc lập với nhau — có thể chọn làm một phần (vd: chỉ observability, bỏ qua AI) tuỳ thời gian, không bắt buộc làm tuần tự hoặc làm hết toàn bộ bảng trên.
