# Quy ước OpenAPI

Mọi file dưới [../../api-docs/openapi/](../../api-docs/openapi/) phải theo đúng quy ước dưới đây — đọc trước khi thêm/sửa bất kỳ endpoint nào. File OpenAPI là **hợp đồng chính thức** giữa các service và giữa backend/frontend (xem [../../AGENTS.md](../../AGENTS.md#nguồn-sự-thật)): sửa hợp đồng trước, code handler sau.

## Bảng chỉ mục

| Service | File | Domain spec liên quan |
|---|---|---|
| Booking Service | [booking-service.yaml](../../api-docs/openapi/booking-service.yaml) | [../02-domains/booking/spec.md](../02-domains/booking/spec.md) |
| Event Service | [event-service.yaml](../../api-docs/openapi/event-service.yaml) | [../02-domains/event-catalog/spec.md](../02-domains/event-catalog/spec.md) |
| Identity Service | [identity-service.yaml](../../api-docs/openapi/identity-service.yaml) | [../02-domains/identity/spec.md](../02-domains/identity/spec.md) |
| Payment Service | [payment-service.yaml](../../api-docs/openapi/payment-service.yaml) | [../02-domains/payment/spec.md](../02-domains/payment/spec.md) |
| Notification Service | [notification-service.yaml](../../api-docs/openapi/notification-service.yaml) | [../02-domains/notification/spec.md](../02-domains/notification/spec.md) |
| File Service | [file-service.yaml](../../api-docs/openapi/file-service.yaml) | [../02-domains/file-storage/spec.md](../02-domains/file-storage/spec.md) |

API Gateway không có file OpenAPI riêng — nó chỉ route/tổng hợp request tới các service trên theo `servers: [{url: /api/v1}]` chung.

## Quy ước bắt buộc

- `openapi: 3.0.3`; `servers: [{url: /api/v1}]`.
- `info.description` mô tả bằng tiếng Việt, trỏ về domain spec liên quan (`docs/02-domains/<domain>/spec.md`) — đặc biệt với luồng có ràng buộc nghiệp vụ phức tạp (transaction, rate limit).
- `security`: khai báo `bearerAuth` (`type: http`, `scheme: bearer`, `bearerFormat: JWT`) ở top-level `security: [{bearerAuth: []}]`; endpoint public (login, refresh, danh sách/chi tiết sự kiện công khai...) override bằng `security: []` ngay trong operation đó — không tạo security scheme thứ hai.
- Lỗi: mọi response không phải 2xx dùng `$ref: '#/components/schemas/Error'`, schema cố định `{code: string, message: string}` — khớp 1:1 với `src/pkg/apperr.Error` (xem [backend-conventions.md](backend-conventions.md#xử-lý-lỗi-pkgapperr)). Không tự định nghĩa schema lỗi khác cho riêng 1 endpoint.
- Phân trang: mọi endpoint trả danh sách dùng chung 2 param `$ref: '#/components/parameters/PageParam'` (`page`, integer, min 1, default 1) và `PageSizeParam` (`page_size`, integer, min 1, max 100, default 20); response là 1 schema riêng theo dạng `<Resource>ListResponse` với đúng 2 field `items` (array) và `total` (integer) — khớp `src/pkg/pagination.Envelope[T]` (field `Page`/`PageSize` là tham số request, không xuất hiện lại trong response).
- Path: danh từ số nhiều cho resource (`/bookings`, `/events`, `/users`), sub-resource dạng động từ cho hành động chuyển trạng thái (`/bookings/{id}/cancel`, `/organizer/events/{id}/publish`). Không dùng verb trong path chính (không viết `/getBookings`).
- Field JSON: snake_case (`ticket_type_id`, `created_at`...).
- **Không dùng `operationId`** ở bất kỳ operation nào trong toàn bộ `api-docs/openapi/` — quy ước có chủ đích của repo, không tự thêm khi viết endpoint mới.
- `tags`: nhóm theo audience/tiền tố path (`bookings`, `organizer`, `admin`, `public`...), không nhóm theo entity kỹ thuật.

## Quy trình thêm/sửa endpoint

1. Xác định domain sở hữu qua [../02-domains/README.md](../02-domains/README.md) và service tương ứng ở bảng chỉ mục trên.
2. Sửa/thêm path trong đúng file `.yaml` của service, theo quy ước ở trên — làm **trước** khi code handler (`AGENTS.md` quy ước #1).
3. Nếu là endpoint hoàn toàn mới, ngoài phạm vi đã ngụ ý trong [system-architecture.md](system-architecture.md), dừng lại và xác nhận với người dùng trước khi thêm.
4. Cập nhật mục "API liên quan" trong domain spec tương ứng nếu endpoint mới đủ quan trọng để nhắc tới ở đó.
5. Parse-check file YAML còn hợp lệ (chưa có tool validate tự động trong repo — đọc lại thủ công cấu trúc `paths`/`components`/`$ref`).
6. Implement handler khớp chính xác path/method/schema vừa định nghĩa (xem [backend-conventions.md](backend-conventions.md)).
