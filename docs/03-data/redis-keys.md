# Redis — key pattern

Redis phục vụ ba mục đích trong hệ thống: **distributed lock**, **cache-aside**, và **rate-limit counter**. Không dùng Redis làm nguồn sự thật cho dữ liệu cần bền vững lâu dài (đơn hàng, vé...) — dữ liệu đó luôn ở PostgreSQL.

## Key pattern

| Key pattern | Mục đích | TTL | Dùng bởi |
|---|---|---|---|
| `lock:ticket_type:{id}` | Distributed lock khi checkout, chặn bớt tải trước khi chạm DB | 5–10s | Booking Service |
| `cache:event:{id}` | Cache chi tiết sự kiện | 5 phút | Event Service |
| `cache:events:trending` | Cache danh sách sự kiện nổi bật | 1 phút | Event Service |
| `ratelimit:{ip}:{route}` | Đếm request cho token bucket | Theo cửa sổ thời gian của route | API Gateway |
| `session:blacklist:{jti}` | Access token bị thu hồi trước hạn (logout) | Bằng thời gian còn lại của token | Identity Service |

## Cache-aside

Đọc cache trước, miss thì đọc DB rồi ghi lại cache. Áp dụng cho danh sách sự kiện nổi bật, chi tiết sự kiện, kết quả search phổ biến. Invalidate cache (`DEL cache:event:{id}`) ngay khi organizer cập nhật sự kiện tương ứng — không chờ TTL hết hạn, tránh hiển thị dữ liệu cũ cho người dùng ngay sau khi organizer sửa.

## Distributed lock

Dùng thư viện kiểu `redsync` (Redlock) để acquire lock theo `ticket_type_id` trước khi vào transaction Postgres khi đặt vé — mục đích giảm tải request đồng thời chạm DB trong kịch bản flash-sale, **không thay thế** `SELECT ... FOR UPDATE` (lock ở Redis chỉ là hàng rào chắn bớt tải ở tầng ứng dụng; tính đúng đắn cuối cùng vẫn do transaction Postgres đảm bảo). Chi tiết luồng xem [../02-domains/booking/spec.md](../02-domains/booking/spec.md).

## Rate-limit counter

Token bucket implement bằng Redis (`INCR` + `EXPIRE`, hoặc Lua script để atomic). Mức giới hạn khác nhau theo route — xem chi tiết tại [../04-security/rate-limiting.md](../04-security/rate-limiting.md).

## Persistence *(nice-to-have, Phase 4)*

Cấu hình kết hợp **RDB** (snapshot định kỳ, phục hồi nhanh) và **AOF** (ghi lại mọi lệnh ghi, mất ít dữ liệu hơn khi crash). Dữ liệu lock/rate-limit tuy ngắn hạn nhưng quan trọng lúc cao điểm (mất lock đang giữ giữa lúc flash-sale có thể dẫn tới race condition tạm thời trước khi Postgres transaction chặn lại) — nên ưu tiên AOF `everysec` thay vì để Redis chạy hoàn toàn in-memory không persistence.
