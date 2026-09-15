# Domain: Identity

**Service sở hữu:** Identity Service · **Database:** PostgreSQL (`users`) · **Phase:** MVP (Phase 1) cho auth cơ bản, Phase 2 cho duyệt organizer + RBAC đầy đủ.

## Phạm vi & trách nhiệm

Quản lý danh tính người dùng: đăng ký/đăng nhập, phát hành và thu hồi token nội bộ, quản lý role (`super_admin`/`organizer`/`user`) và trạng thái tài khoản (`active`/`pending`/`banned`). Đây là service duy nhất được ghi vào bảng `users` — mọi service khác chỉ đọc thông tin user qua gRPC (vd: Event Service xác thực `organizer_id` tồn tại), không truy vấn trực tiếp database của Identity Service.

## Data model

Bảng `users` — định nghĩa đầy đủ tại [../../03-data/postgres-schema.md](../../03-data/postgres-schema.md#identity-service--users).

## Luồng nghiệp vụ

### Đăng nhập / đăng ký lần đầu

1. Client đăng nhập qua Firebase SDK, nhận Firebase ID token.
2. `POST /auth/login` với ID token → Identity Service verify bằng Firebase Admin SDK.
3. Nếu `firebase_uid` chưa tồn tại trong `users`: tạo user mới với `role='user'`, `status='active'`. Nếu đã tồn tại: tra thông tin hiện có.
4. Phát `access_token` (JWT 15 phút) + `refresh_token` (cookie httpOnly 7 ngày).

Ở môi trường non-production, Identity Service có thể chạy với một mock Firebase verifier (cấu hình qua biến môi trường, không cần Firebase project thật) — verifier này chấp nhận một token tự tạo cục bộ thay cho Firebase ID token thật ở bước 2, nhưng vẫn phát hành `access_token`/`refresh_token` theo đúng hợp đồng như trên.

Chi tiết đầy đủ về token, claim, rotation, blacklist xem [../../04-security/authentication.md](../../04-security/authentication.md).

### Đăng ký trở thành organizer

1. User (role hiện tại `user`) gọi endpoint yêu cầu trở thành organizer → tài khoản chuyển `status='pending'` (giữ nguyên `role='user'` cho tới khi được duyệt — **không** tự nâng `role='organizer'` ngay).
2. `super_admin` xem danh sách yêu cầu đang `pending` qua `GET /admin/users` (filter theo status), duyệt bằng `PATCH /admin/users/:id/role` → cập nhật `role='organizer'`, `status='active'`.
3. Hành động duyệt được ghi vào `audit_logs` (xem [../../04-security/authorization.md](../../04-security/authorization.md)).
4. Nếu bị từ chối: `status` quay lại `active` với `role='user'` (không có "banned" cho trường hợp này — từ chối không phải khoá tài khoản).

### Khoá / mở khoá tài khoản

Chỉ `super_admin`. `PATCH /admin/users/:id/role` (hoặc endpoint riêng nếu tách bạch role-change và status-change) cập nhật `status='banned'`. Tài khoản `banned` bị từ chối ngay ở bước xác thực JWT tại API Gateway (kiểm tra `status` hiện tại của user, không chỉ tin vào claim trong token đã phát trước đó — token cũ có thể còn hạn nhưng user đã bị khoá sau khi phát hành).

## Quan hệ với domain khác

- **event-catalog**: `events.organizer_id` tham chiếu `users.id`; Event Service xác thực organizer tồn tại và có `role='organizer'` trước khi cho tạo sự kiện.
- **booking**: `orders.user_id` tham chiếu `users.id`.
- Mọi domain đều dựa vào claim `role` trong JWT do Identity Service phát hành để chạy middleware `RequireRole` (xem [../../04-security/authorization.md](../../04-security/authorization.md)).

## API liên quan

Xem đặc tả đầy đủ tại [../../../api-docs/openapi/identity-service.yaml](../../../api-docs/openapi/identity-service.yaml).

## Phân theo phase

| Tính năng | Phase |
|---|---|
| Đăng nhập/đăng xuất, JWT, refresh token rotation | 1 (MVP) |
| `GET/PATCH /users/me` | 1 (MVP) |
| Duyệt/khoá tài khoản, RBAC đầy đủ, audit log | 2 (Must-have) |
