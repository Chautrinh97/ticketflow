# Background job / Cron job

**Phase 2 (docker-compose):** mỗi job là một binary/container riêng chạy scheduler trong-process (vd: `robfig/cron` trong Go) bên trong `deployments/docker-compose.yaml`, không chạy cron lồng trong cùng tiến trình với service phục vụ request (service chính vẫn chỉ phục vụ HTTP/gRPC). Lý do: Kubernetes chưa triển khai tới Phase 3 (xem [deployment.md](deployment.md)), nên chưa thể dùng `CronJob` thật.

**Phase 3+ (Kubernetes):** chuyển từng job trên thành một **Kubernetes CronJob** riêng, có thể scale/monitor độc lập với service chính — giữ nguyên logic nghiệp vụ, chỉ đổi cơ chế lập lịch.

## Danh sách job

| Job | Lịch chạy | Việc làm | Thuộc domain | Phase |
|---|---|---|---|---|
| Huỷ đơn hàng quá hạn | Mỗi 1 phút | Tìm `orders` có `status='pending'` và `expires_at < now()` → hoàn `sold_count` trong `ticket_types`, cập nhật `orders.status='expired'` | [../02-domains/booking/spec.md](../02-domains/booking/spec.md) | 2 |
| Nhắc lịch sự kiện | Mỗi giờ | Tìm sự kiện diễn ra trong 24h tới → publish `event.reminder` | [../02-domains/notification/spec.md](../02-domains/notification/spec.md) | 2 (cần Notification Service) |
| Báo cáo doanh thu ngày | 0h hằng ngày | Tổng hợp doanh thu theo organizer, lưu snapshot | [../02-domains/analytics/spec.md](../02-domains/analytics/spec.md) | 2 (cần Analytics) |
| Đồng bộ Elasticsearch *(nice-to-have)* | Theo CDC hoặc mỗi 5 phút | Đồng bộ thay đổi từ Postgres/Mongo sang index tìm kiếm | [../02-domains/search/spec.md](../02-domains/search/spec.md) | 2+ (nice-to-have) |

## Job quan trọng nhất: huỷ đơn hàng quá hạn

Đây là job có tác động trực tiếp tới tính đúng đắn của tồn kho vé — nếu job chạy chậm hoặc lỗi, vé bị "giữ ảo" bởi các đơn hàng đã hết hạn thanh toán mà chưa được hoàn lại, làm giảm số vé thực tế còn bán được. Job phải:

- Chạy trong transaction giống hệt nguyên tắc ở luồng đặt vé (`SELECT ... FOR UPDATE` trên `ticket_types` trước khi cộng lại `sold_count`) — tránh race condition với một request đặt vé mới đang chạy song song trên cùng `ticket_type_id`.
- Idempotent: chạy lại nhiều lần trên cùng một `order_id` đã `expired` không được trừ/cộng `sold_count` thêm lần nữa.
- Có alerting nếu job không chạy được liên tục quá X phút (X cụ thể xác định khi implement observability, xem [observability.md](observability.md)).
