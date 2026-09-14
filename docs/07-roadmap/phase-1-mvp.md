# Phase 1 — MVP

**Ước lượng:** 2–3 tuần. **Mục tiêu:** một luồng end-to-end chạy được — người dùng xem sự kiện, đặt vé, hệ thống không bán trùng vé — chạy bằng `docker-compose`, chưa cần Kubernetes.

## Phạm vi theo domain

| Domain | Việc cần làm trong Phase 1 |
|---|---|
| [identity](../02-domains/identity/spec.md) | Đăng ký/đăng nhập qua Firebase, phát JWT nội bộ, `GET/PATCH /users/me`. Chưa cần duyệt organizer thủ công phức tạp — có thể seed sẵn vài tài khoản `organizer`/`super_admin` để phát triển |
| [event-catalog](../02-domains/event-catalog/spec.md) | CRUD sự kiện cơ bản (không cần validate `attributes` theo category chặt chẽ), tạo `ticket_types`, `GET /events`, `GET /events/:slug` |
| [booking](../02-domains/booking/spec.md) | `POST /bookings` với transaction ACID (`SELECT ... FOR UPDATE`) — đây là phần **bắt buộc làm đúng ngay từ đầu**, không phải phần có thể làm tạm rồi sửa sau |
| [payment](../02-domains/payment/spec.md) | Có thể mock/giả lập cổng thanh toán (endpoint nội bộ tự đánh dấu thành công) thay vì tích hợp gateway thật — miễn luồng `orders.status: pending → paid` hoạt động đúng |
| [notification](../02-domains/notification/spec.md) | Chưa cần — có thể hoãn sang Phase 2 |
| [file-storage](../02-domains/file-storage/spec.md) | Chưa cần — banner sự kiện có thể dùng URL nhập tay |

## Frontend

Next.js: trang danh sách sự kiện, chi tiết sự kiện, luồng đặt vé cơ bản (CSR, có auth). Chưa cần dashboard organizer/admin đầy đủ.

## API & docs

REST đầy đủ cho 3 domain trên, có Swagger UI chạy được (sinh từ swaggo hoặc dùng thẳng OpenAPI tại [../../api-docs/openapi/](../../api-docs/openapi/)).

## Ngoài phạm vi Phase 1 (để dành Phase 2+)

RBAC đầy đủ theo ma trận quyền (Phase 1 chỉ cần phân biệt được authenticated/organizer/admin ở mức tối thiểu để luồng chạy được), full-text/fuzzy search, Redis cache, rate limit, cron huỷ đơn quá hạn, mọi thứ ở Phase 3/4.
