# Bảo mật & Rate limiting

## Các lớp bảo vệ

| Lớp bảo vệ | Cơ chế |
|---|---|
| SQL Injection | Parameterized query qua ORM/`pgx` — không nối chuỗi SQL thủ công |
| Rate limit theo endpoint thường | 100 request/phút/IP hoặc user (token bucket, state lưu Redis) |
| Rate limit endpoint đặt vé | 5 request/10 giây/user — chống bot mua vé hàng loạt, siết chặt hơn hẳn mức thường |
| CORS | Whitelist đúng domain frontend, `credentials: true` |
| CSRF | Vì refresh token lưu ở cookie → thêm CSRF token cho các request state-changing (POST/PUT/PATCH/DELETE) |
| DoS tầng ứng dụng | Giới hạn kích thước request body, timeout kết nối, circuit breaker giữa các service |
| DoS/DDoS tầng edge | Cloudflare WAF + rate-limit rule + "I'm Under Attack Mode" khi cần |
| Webhook thanh toán | Xác thực chữ ký HMAC từ payment gateway, xử lý idempotent theo `provider_txn_id` |

## Rate limit — nơi thực thi

Thực thi ở **API Gateway** (lớp đầu tiên chạm mọi request), dùng thuật toán token bucket với counter lưu ở Redis (`ratelimit:{ip}:{route}`, xem [../03-data/redis-keys.md](../03-data/redis-keys.md)). Vượt hạn mức trả về `429 Too Many Requests`. Cloudflare ở tầng edge là lớp phòng vệ bổ sung phía trước API Gateway, không thay thế rate-limit ở tầng ứng dụng (Cloudflare chặn theo IP/pattern chung; API Gateway chặn theo user/route cụ thể, hiểu ngữ cảnh nghiệp vụ).

## Vì sao endpoint đặt vé cần rate limit riêng, chặt hơn

Đây là endpoint có giá trị kinh tế cao nhất (mua được vé = có giá trị bán lại) và dễ bị bot tấn công nhất trong kịch bản flash-sale. Giới hạn 5 request/10 giây/user là hàng rào đầu tiên; hàng rào thứ hai là distributed lock theo `ticket_type_id` (xem [../03-data/redis-keys.md](../03-data/redis-keys.md)); hàng rào cuối cùng, mang tính quyết định đúng-sai, là transaction Postgres với `SELECT ... FOR UPDATE` (xem [../02-domains/booking/spec.md](../02-domains/booking/spec.md)).

## Webhook thanh toán — xác thực & idempotency

- **Xác thực chữ ký**: mọi request tới `POST /payments/webhook` phải verify chữ ký HMAC do payment gateway ký (secret key trao đổi trước với gateway) — từ chối ngay nếu chữ ký sai, không xử lý payload.
- **Idempotent theo `provider_txn_id`**: gateway có thể gọi lại cùng một webhook nhiều lần (retry). Payment Service phải kiểm tra `provider_txn_id` đã tồn tại trong bảng `payments` với `status` cuối cùng chưa — nếu đã xử lý xong (thành công hoặc thất bại), trả về `200 OK` ngay mà không xử lý lại logic nghiệp vụ (không publish lại event, không cộng/trừ tồn kho lần hai).

Chi tiết luồng đầy đủ xem [../02-domains/payment/spec.md](../02-domains/payment/spec.md).
