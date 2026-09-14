# Lộ trình triển khai — tổng quan

| Giai đoạn | Nội dung | Ước lượng |
|---|---|---|
| [Phase 1 — MVP](phase-1-mvp.md) | Identity + Event + Booking (transaction ACID) + REST + Swagger + Next.js xem/đặt vé, chạy bằng docker-compose | 2–3 tuần |
| [Phase 2 — Must-have hoàn chỉnh](phase-2-must-have.md) | RBAC đầy đủ, full-text/fuzzy search, upload file, Redis cache, rate limit, cron huỷ đơn, notification email | 2–3 tuần |
| [Phase 3 — Hạ tầng production](phase-3-infra.md) | Deploy Kubernetes (Helm), Ingress + cert-manager, Cloudflare, HPA, health check, CI/CD, demo bằng k9s | 1–2 tuần |
| [Phase 4 — Nice-to-have](phase-4-nice-to-have.md) | RabbitMQ/Kafka + saga, OpenTelemetry + Grafana, Elasticsearch, Redis AOF/RDB, tính năng AI | 2–3 tuần |

> Dừng ở Phase 2 hoặc 3 vẫn là một portfolio chắc chắn; Phase 4 dành cho lúc có thời gian mở rộng thêm.

## Nguyên tắc chọn phạm vi mỗi phase

- Mỗi phase phải **chạy được end-to-end** (dù còn thô) trước khi chuyển sang phase sau — không để dở dang nhiều phase cùng lúc.
- Tính năng đánh dấu *(nice-to-have)* trong các domain spec ([../02-domains/](../02-domains/)) mặc định thuộc Phase 4, trừ khi ghi chú khác.
- Khi implement, đối chiếu domain spec + phase doc tương ứng để biết chính xác nên làm gì trong phase hiện tại, tránh làm vượt phạm vi (gây trì hoãn phase) hoặc thiếu phạm vi (phase không chạy được end-to-end).
