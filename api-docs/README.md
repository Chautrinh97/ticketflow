# API Docs — TicketFlow

Hợp đồng API chính thức (OpenAPI 3.0) giữa client (frontend/mobile) và hệ thống, theo mô hình **contract-first**: file OpenAPI được viết trước khi có code, dùng làm cơ sở để frontend và backend phát triển song song. Khi implement, handler backend phải khớp với các file này; nếu cần đổi hợp đồng, sửa OpenAPI trước rồi mới sửa code (xem quy ước tại [../AGENTS.md](../AGENTS.md)).

## Vị trí các file

```
api-docs/openapi/
├── identity-service.yaml
├── event-service.yaml
├── booking-service.yaml
├── payment-service.yaml
├── notification-service.yaml
└── file-service.yaml
```

Không có file riêng cho **API Gateway** — Gateway chỉ forward request theo path prefix tới đúng service phía sau, giữ nguyên path và request/response (`/auth/*`, `/users/me`, `/users/me/organizer-request`, `/admin/*` → Identity; `/events/*`, `/organizer/events/*` → Event; `/bookings/*`, `/users/me/bookings` → Booking; `/payments/*` → Payment; `/files/*` → File; `/users/me/notifications/*` → Notification), tất cả dưới tiền tố chung `/api/v1`. Lưu ý `/users/me/bookings` thuộc Booking Service dù chung tiền tố `/users/me` với Identity Service — route theo endpoint cụ thể, không chỉ theo tiền tố `/users/*`.

## Quy ước

- Mỗi service một file YAML độc lập, tự chứa toàn bộ `components.schemas` nó cần (không `$ref` chéo giữa các file) — vì mỗi service có thể được xem/generate độc lập.
- Security scheme dùng chung tên `bearerAuth` (JWT, header `Authorization: Bearer <token>`) ở mọi file — endpoint không cần auth khai báo `security: []` để override.
- Response lỗi dùng chung shape `Error` (`code`, `message`) khai báo trong từng file.
- Version API nằm trong path (`/api/v1`), không dùng header versioning.

## Cách xem

Xem trực tiếp file YAML hoặc paste vào [Swagger Editor](https://editor.swagger.io/). File tại `api-docs/` là nguồn thiết kế chính thức — mỗi service phía sau chạy REST API khớp đúng file tương ứng của nó, xác nhận bằng smoke test (`scripts/smoke-test.sh`) và test tích hợp qua `docker compose up`.

## Domain spec tương ứng

Mỗi endpoint gắn với một domain nghiệp vụ mô tả chi tiết tại [../docs/02-domains/](../docs/02-domains/) — đọc domain spec trước khi implement handler cho endpoint.
