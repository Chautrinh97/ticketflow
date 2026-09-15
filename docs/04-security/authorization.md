# Authorization

## Hai lớp kiểm tra bắt buộc

Mọi endpoint thao tác lên resource cụ thể phải qua **hai lớp kiểm tra**, không được chỉ dừng ở lớp thứ nhất:

1. **Role-based**: user hiện tại có `role` phù hợp để gọi endpoint này không (vd: chỉ `organizer` hoặc `super_admin` mới gọi được `POST /organizer/events`).
2. **Ownership-based**: nếu endpoint thao tác lên một resource cụ thể đã tồn tại (sửa/xoá sự kiện, huỷ đơn hàng...), user hiện tại có phải chủ sở hữu của resource đó không — **trừ khi** là `super_admin` (được bỏ qua kiểm tra ownership theo ma trận quyền, xem [../00-overview/roles-permissions.md](../00-overview/roles-permissions.md)).

Thiếu lớp (2) là lỗi bảo mật phổ biến nhất trong hệ thống dạng này: một `organizer` hợp lệ (qua được kiểm tra role) vẫn có thể sửa/xoá sự kiện của `organizer` khác nếu chỉ dừng ở kiểm tra role.

## Middleware pattern (Go, minh hoạ chữ ký hàm)

```go
// RequireRole chỉ cho phép user có role nằm trong danh sách truyền vào.
func RequireRole(roles ...string) gin.HandlerFunc { ... }

// RequireOwnership so khớp owner của resource (lấy qua resourceLookup) với user hiện tại.
// super_admin luôn được bỏ qua kiểm tra này.
func RequireOwnership(resourceLookup func(ctx *gin.Context) (ownerID string, err error)) gin.HandlerFunc { ... }
```

`resourceLookup` phải truy vấn resource thật (vd: `SELECT organizer_id FROM events WHERE id = ?`) — không được suy luận owner từ input do client gửi lên (client có thể gửi sai `organizer_id` để giả mạo).

## Áp dụng theo từng nhóm endpoint

| Nhóm endpoint | Role yêu cầu | Kiểm tra ownership? |
|---|---|---|
| `POST /organizer/events` | `organizer` | Không (resource chưa tồn tại — owner = user hiện tại) |
| `PATCH/DELETE /organizer/events/:id` | `organizer` (hoặc `super_admin`) | Có — `event.organizer_id == current_user.id`, bỏ qua nếu `super_admin` |
| `POST /bookings/:id/cancel` | `user` (chủ đơn) hoặc `super_admin` | Có — `order.user_id == current_user.id`, bỏ qua nếu `super_admin` |
| `GET /organizer/events`, `GET /organizer/events/{eventId}/buyers`, `GET /organizer/stats`, `GET /organizer/events/{id}/stats` | `organizer` | Có với các endpoint theo 1 sự kiện cụ thể (`buyers`, `events/{id}/stats`) — `event.organizer_id == current_user.id`; các endpoint tổng hợp (`GET /organizer/events`, `GET /organizer/stats`) tự giới hạn theo `current_user.id`, không cần tham số ownership riêng |
| `GET /admin/users`, `PATCH /admin/users/:id/role`, `GET /admin/stats`, `GET /admin/audit-logs` | `super_admin` | Không áp dụng (không có khái niệm "chủ sở hữu" cho dữ liệu toàn hệ thống) |

## Audit log

Các hành động nhạy cảm thực hiện bởi `super_admin` (khoá/mở khoá tài khoản, duyệt organizer, đổi role, huỷ vé thay người dùng) phải ghi vào bảng `audit_logs` (xem [../03-data/postgres-schema.md](../03-data/postgres-schema.md)) — ghi ngay trong cùng transaction hoặc ngay sau khi hành động thành công, không ghi bất đồng bộ qua queue (tránh mất log nếu queue lỗi).
