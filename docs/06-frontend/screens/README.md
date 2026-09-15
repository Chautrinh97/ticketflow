# Screens — chỉ mục màn hình

Mỗi màn hình có 1 file spec riêng, viết theo khuôn mẫu ở mục "Khuôn mẫu 1 file screen spec" bên dưới. Trước khi đọc chi tiết 1 màn hình, đọc qua [../design-system.md](../design-system.md), [../interaction-patterns.md](../interaction-patterns.md), và [../components/](../components/) — mọi screen spec đều giả định người đọc đã biết các quy ước đó và chỉ tham chiếu lại, không lặp.

## Bảng chỉ mục

| Màn hình | Vai trò | Phase | Route gợi ý | API chính |
|---|---|---|---|---|
| [public/auth-login.md](public/auth-login.md) | public | 1 | `/login` | `POST /auth/login` |
| [public/home.md](public/home.md) | public | 1 | `/` | `GET /events` |
| [public/event-search.md](public/event-search.md) | public | 1 (filter cơ bản) → 2 (full-text/fuzzy) | `/events` | `GET /events`, `GET /events/search` |
| [public/event-detail.md](public/event-detail.md) | public | 1 | `/events/[slug]` | `GET /events/:slug` |
| [user/checkout.md](user/checkout.md) | user | 1 | `/events/[slug]/checkout` | `POST /bookings`, `POST /payments/:orderId/checkout` |
| [user/order-confirmation.md](user/order-confirmation.md) | user | 1 | `/orders/[orderId]/confirmation` | `GET /bookings/:id` |
| [user/my-bookings.md](user/my-bookings.md) | user | 1 | `/me/bookings` | `GET /users/me/bookings` |
| [user/booking-detail.md](user/booking-detail.md) | user | 1 (xem) → 2 (huỷ vé) | `/me/bookings/[id]` | `GET /bookings/:id`, `POST /bookings/:id/cancel` |
| [user/profile.md](user/profile.md) | user | 1 | `/me/profile` | `GET/PATCH /users/me` |
| [user/organizer-request.md](user/organizer-request.md) | user | 2 | `/me/organizer-request` | `POST /users/me/organizer-request` |
| [user/notifications.md](user/notifications.md) | user | 2 | `/me/notifications` | `GET /users/me/notifications`, `PATCH .../read` |
| [organizer/dashboard.md](organizer/dashboard.md) | organizer | 2 | `/organizer` | `GET /organizer/stats` |
| [organizer/event-list.md](organizer/event-list.md) | organizer | 2 | `/organizer/events` | `GET /organizer/events` |
| [organizer/event-form.md](organizer/event-form.md) | organizer | 1 (tạo cơ bản) → 2 (đầy đủ) | `/organizer/events/new`, `/organizer/events/[id]/edit` | `POST /organizer/events`, `PATCH /organizer/events/:id`, `POST .../ticket-types`, `POST /files/presign` |
| [organizer/event-manage.md](organizer/event-manage.md) | organizer | 1 (cơ bản) → 2 (đủ tab) | `/organizer/events/[id]` | `POST .../publish`, `GET .../buyers`, `GET .../stats` |
| [admin/dashboard.md](admin/dashboard.md) | admin | 2 | `/admin` | `GET /admin/stats` |
| [admin/user-management.md](admin/user-management.md) | admin | 2 | `/admin/users` | `GET /admin/users`, `PATCH /admin/users/:id/role` |
| [admin/audit-log.md](admin/audit-log.md) | admin | 2 | `/admin/audit-logs` | `GET /admin/audit-logs` |
| [system/error-states.md](system/error-states.md) | mọi vai trò | 1 | — | không gọi API riêng |

## Khuôn mẫu 1 file screen spec

Mọi file trong `screens/**/*.md` viết theo đúng 6 mục sau, theo đúng thứ tự:

1. **Header**: tên màn hình, route, vai trò truy cập, render mode (tham chiếu bảng tại [../README.md](../README.md)), phase, link domain spec + OpenAPI liên quan.
2. **Layout tổng quan**: sơ đồ khối ASCII thể hiện vị trí section (top/left/center/right/bottom) + bảng liệt kê section — ghi rõ section nào là **component dùng chung** (link `../components/*.md`) vs section **riêng của màn hình này**.
3. **Chi tiết từng section riêng**: unit con, nội dung hiển thị, field dữ liệu nguồn (map field trong response OpenAPI), màu sắc/style riêng dùng token từ [../design-system.md](../design-system.md). Section đã là component dùng chung thì **không** mô tả lại — chỉ nêu tên + variant áp dụng.
4. **Hành vi tương tác**: bảng "Hành động người dùng → Phản hồi hệ thống". Quy tắc chung (khi nào toast/modal/loading...) tham chiếu [../interaction-patterns.md](../interaction-patterns.md), chỉ nêu nội dung/text cụ thể của màn hình.
5. **Trạng thái đặc biệt**: nội dung cụ thể cho empty/loading/error của màn hình này (cấu trúc chung xem `interaction-patterns.md`).
6. **Responsive**: khác biệt đáng kể ở mobile, nếu có.

## Quy trình thêm màn hình mới

1. Xác định vai trò truy cập, phase, route, và các endpoint API sẽ dùng.
2. Rà soát [../components/](../components/) — thành phần cần dùng đã có chưa:
   - Có rồi → chỉ tham chiếu tên + variant trong screen spec mới, không mô tả lại.
   - Chưa có → thêm vào `components/*.md` phù hợp **trước** (theo quy trình tại [../components/README.md](../components/README.md#quy-trình-thêm-component-mới)), rồi mới viết screen tham chiếu tới nó.
3. Nếu cần endpoint API chưa tồn tại: cập nhật `api-docs/openapi/` trước (đúng convention hiện có: schema, `security`, response `Error` chung), rồi mới viết screen tham chiếu endpoint đó — không mô tả UI dựa trên API chưa được định nghĩa.
4. Viết file screen mới theo đúng "Khuôn mẫu 1 file screen spec" ở trên.
5. Thêm 1 dòng vào "Bảng chỉ mục" ở đầu file này (màn hình, vai trò, phase, route, API chính).
