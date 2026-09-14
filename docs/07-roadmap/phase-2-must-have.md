# Phase 2 — Must-have hoàn chỉnh

**Ước lượng:** 2–3 tuần. **Mục tiêu:** hệ thống có đủ các cơ chế "phải có" của một sản phẩm thật — phân quyền chặt chẽ, tìm kiếm tốt, thông báo, bảo vệ khỏi lạm dụng — vẫn chạy bằng `docker-compose`.

## Phạm vi theo domain/mối quan tâm

| Hạng mục | Việc cần làm | Tài liệu liên quan |
|---|---|---|
| RBAC đầy đủ | Áp dụng đầy đủ ma trận quyền, middleware `RequireRole` + `RequireOwnership` trên mọi endpoint liên quan | [../00-overview/roles-permissions.md](../00-overview/roles-permissions.md), [../04-security/authorization.md](../04-security/authorization.md) |
| Duyệt organizer | Luồng đăng ký organizer → `status='pending'` → `super_admin` duyệt | [../02-domains/identity/spec.md](../02-domains/identity/spec.md) |
| Full-text + fuzzy search | `tsvector`/`tsquery` + `pg_trgm` trên Postgres | [../02-domains/event-catalog/spec.md](../02-domains/event-catalog/spec.md), [../03-data/postgres-schema.md](../03-data/postgres-schema.md) |
| Upload file | File Service sinh presigned URL, banner sự kiện/avatar dùng URL thật từ object storage | [../02-domains/file-storage/spec.md](../02-domains/file-storage/spec.md) |
| Redis cache | Cache-aside cho chi tiết/danh sách sự kiện, invalidate khi organizer cập nhật | [../03-data/redis-keys.md](../03-data/redis-keys.md) |
| Rate limit | Token bucket ở API Gateway, siết chặt riêng endpoint đặt vé | [../04-security/rate-limiting.md](../04-security/rate-limiting.md) |
| Cron huỷ đơn quá hạn | Job huỷ `orders` pending quá `expires_at`, hoàn `sold_count` | [../05-infra-devops/background-jobs.md](../05-infra-devops/background-jobs.md), [../02-domains/booking/spec.md](../02-domains/booking/spec.md) |
| Notification email | Gửi email xác nhận đơn/thanh toán/huỷ vé qua SendGrid/SES, ghi bảng `notifications` | [../02-domains/notification/spec.md](../02-domains/notification/spec.md) |
| Thanh toán thật | Tích hợp cổng thanh toán thật (Stripe/VNPay/Momo), xử lý webhook có xác thực chữ ký + idempotent | [../02-domains/payment/spec.md](../02-domains/payment/spec.md) |
| Thống kê organizer | `GET` báo cáo doanh thu/tỉ lệ bán vé theo sự kiện của organizer | [../02-domains/analytics/spec.md](../02-domains/analytics/spec.md) |

## Ngoài phạm vi Phase 2

Triển khai Kubernetes/CI-CD thật (Phase 3), message queue/saga/observability/Elasticsearch/AI (Phase 4).
