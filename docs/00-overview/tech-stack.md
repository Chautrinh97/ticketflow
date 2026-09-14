# Tech stack

| Nhóm | Công nghệ | Ghi chú |
|---|---|---|
| Ngôn ngữ backend | Go 1.22+ | |
| HTTP framework | Gin (hoặc Echo/Fiber) | Middleware chain: recovery → logging → CORS → rate-limit → auth |
| Giao tiếp nội bộ | gRPC + Protocol Buffers | Interceptor cho auth, logging, trace-id propagation |
| ORM/Query | GORM hoặc sqlc + `pgx` | Booking Service dùng raw SQL + transaction tường minh để kiểm soát `SELECT ... FOR UPDATE` |
| SQL Database | PostgreSQL 15+ | Users, orders, tickets, payments — dữ liệu giao dịch chặt chẽ |
| NoSQL Database | MongoDB 6+ | Thuộc tính linh hoạt theo loại sự kiện |
| Cache/Lock | Redis 7 | Cache-aside, distributed lock (redsync), rate limit counter |
| Search *(nice-to-have)* | Elasticsearch 8 | Full-text + fuzzy + autocomplete, tách khỏi DB giao dịch |
| Message queue *(nice-to-have)* | RabbitMQ (hoặc Kafka) | Event-driven pipeline, saga pattern |
| Auth | Firebase Authentication + JWT tự phát hành | Access token 15 phút, refresh token 7 ngày |
| Object storage | DigitalOcean Spaces / AWS S3 (S3-compatible) | Presigned URL upload |
| API docs | swaggo (sinh từ code annotation) song song với OpenAPI contract-first tại `api-docs/` | |
| Frontend | Next.js 14 (App Router) + Tailwind CSS | React Server Components cho trang SEO, TanStack Query cho data fetching |
| Container | Docker (multi-stage build) | |
| Orchestration | Kubernetes + Helm, quản lý bằng k9s | HPA, liveness/readiness probe |
| Ingress/SSL | Nginx/Traefik Ingress + cert-manager (Let's Encrypt) | |
| Edge/CDN | Cloudflare | CDN, WAF, DDoS protection, rate-limit biên |
| Observability *(nice-to-have)* | OpenTelemetry + Grafana Tempo/Loki + Prometheus | |
| AI *(nice-to-have)* | OpenAI/Claude API (inference-only) + pgvector | Không train/fine-tune model |
| CI/CD | GitHub Actions | lint → test → build → push image → deploy |

## Nguyên tắc chọn công nghệ

- **Database-per-service**: mỗi service sở hữu database riêng, không share schema trực tiếp giữa các service — xem [../01-architecture/system-architecture.md](../01-architecture/system-architecture.md).
- **PostgreSQL cho dữ liệu cần ACID** (đơn hàng, vé, thanh toán); **MongoDB cho thuộc tính linh hoạt** không cần transaction chặt (catalog sự kiện mở rộng); **Redis cho dữ liệu ngắn hạn** (lock, cache, rate-limit counter).
- Các mục đánh dấu *(nice-to-have)* thuộc Phase 4 — xem [../07-roadmap/phase-4-nice-to-have.md](../07-roadmap/phase-4-nice-to-have.md); ở MVP, search dùng `tsvector`/`pg_trgm` của Postgres thay vì Elasticsearch, và giao tiếp giữa Booking/Payment/Notification có thể tạm dùng gọi trực tiếp trước khi có message queue.
