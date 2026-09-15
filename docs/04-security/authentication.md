# Authentication

## Luồng đăng nhập

1. Client đăng nhập qua Firebase SDK (Google/Facebook/email-password) → nhận **Firebase ID token**.
2. Client gửi ID token này tới `POST /auth/login`.
3. Backend (Identity Service) dùng **Firebase Admin SDK** verify token → tra hoặc khởi tạo user tương ứng trong PostgreSQL (`users`, xem [../03-data/postgres-schema.md](../03-data/postgres-schema.md)).
4. Backend phát cặp token nội bộ:
   - **`access_token`** (JWT, hạn 15 phút) — gửi qua header `Authorization: Bearer <token>`.
   - **`refresh_token`** (hạn 7 ngày) — lưu trong cookie `httpOnly + Secure + SameSite=Strict`, không bao giờ gửi qua response body hay lưu ở localStorage.
5. `POST /auth/refresh` dùng refresh token trong cookie để cấp access token mới. Áp dụng **refresh token rotation**: refresh token là giá trị ngẫu nhiên dùng trực tiếp làm khoá Redis `refresh:{token}` (giá trị là `user_id`, TTL 7 ngày — xem [../03-data/redis-keys.md](../03-data/redis-keys.md)); mỗi lần refresh, token cũ bị xoá ngay khi đọc (get-then-delete) và một token mới được cấp — nếu cùng một refresh token cũ bị dùng lại lần 2 sẽ không tìm thấy key (đã bị xoá), coi như refresh thất bại.
6. **Logout**: xoá cookie refresh token phía client + đưa `jti` (JWT ID) của access token hiện tại vào `session:blacklist:{jti}` trên Redis cho tới khi access token hết hạn tự nhiên (xem [../03-data/redis-keys.md](../03-data/redis-keys.md)).

## Vì sao không dùng thẳng Firebase ID token cho toàn bộ hệ thống

Firebase ID token có TTL và claim cố định theo Firebase, không mang được `role`/`status` nội bộ (super_admin/organizer/user, banned...) và không có cơ chế thu hồi tức thời phù hợp với nhu cầu backend (blacklist, rotation). JWT tự phát hành cho phép nhúng đúng claim cần thiết, kiểm soát TTL ngắn cho access token, và tách biệt vòng đời auth của bên thứ ba (Firebase) khỏi vòng đời phiên làm việc trong hệ thống.

## Claim JWT tối thiểu (access token)

| Claim | Ý nghĩa |
|---|---|
| `sub` | `user.id` (UUID nội bộ) |
| `role` | `super_admin` \| `organizer` \| `user` — dùng cho middleware `RequireRole`, xem [authorization.md](authorization.md) |
| `jti` | ID duy nhất của token, dùng để blacklist khi logout |
| `exp` | Thời điểm hết hạn (15 phút sau khi phát hành) |

## Endpoint liên quan

Xem đặc tả đầy đủ (request/response) tại [api-docs/openapi/identity-service.yaml](../../api-docs/openapi/identity-service.yaml) và spec nghiệp vụ tại [../02-domains/identity/spec.md](../02-domains/identity/spec.md).
