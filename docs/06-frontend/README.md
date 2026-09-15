# Frontend (Next.js)

Tổng quan kỹ thuật frontend TicketFlow. Đặc tả chi tiết từng màn hình và component dùng chung nằm ở các thư mục con — xem mục lục cuối file.

## Render mode theo trang

| Trang | Render mode | Lý do |
|---|---|---|
| [Đăng nhập/Đăng ký](screens/public/auth-login.md) | CSR | Tương tác Firebase SDK, không cần SEO |
| [Trang chủ](screens/public/home.md) | ISR | Cập nhật định kỳ, tốt cho SEO, không cần dữ liệu real-time tuyệt đối |
| [Danh sách/tìm kiếm sự kiện](screens/public/event-search.md) | SSR (tải đầu theo query URL) + CSR (đổi filter) | Link chia sẻ/SEO đúng kết quả, tương tác filter mượt sau đó |
| [Chi tiết sự kiện](screens/public/event-detail.md) | SSR | Cần dữ liệu mới nhất (tồn kho vé còn lại) tại thời điểm request |
| [Đặt vé/checkout](screens/user/checkout.md) | CSR | Tương tác nhiều bước, cần auth, không cần SEO |
| [Kết quả thanh toán](screens/user/order-confirmation.md) | CSR | Cần polling trạng thái, cần auth |
| [Vé của tôi](screens/user/my-bookings.md), [Chi tiết vé](screens/user/booking-detail.md) | CSR (cần auth) | Dữ liệu cá nhân hoá, không cần SEO |
| [Thông tin cá nhân](screens/user/profile.md), [Đăng ký Organizer](screens/user/organizer-request.md) (Phase 2), [Thông báo](screens/user/notifications.md) (Phase 2) | CSR (cần auth) | Dữ liệu cá nhân hoá |
| Dashboard [organizer](screens/organizer/dashboard.md)/[admin](screens/admin/dashboard.md) (Phase 2) và mọi trang quản trị bên dưới | CSR (cần auth) | Dữ liệu cá nhân hoá, tương tác nhiều |
| [Trạng thái lỗi](screens/system/error-states.md) | 404: SSR (trả đúng status code) · lỗi runtime/mất mạng: CSR | SEO/crawler cho 404, còn lại là lỗi phía client |

## Chi tiết kỹ thuật

- **Next.js App Router**, dùng React Server Components cho các trang cần SEO (trang chủ, chi tiết sự kiện) để giảm JS gửi xuống client.
- **TanStack Query (React Query)** quản lý cache dữ liệu phía client cho các trang CSR — đặc biệt cần cho trạng thái tồn kho vé (`ticket_types.sold_count`/`quota`) có thể thay đổi trong lúc người dùng đang ở trang checkout.
- **Tailwind CSS** cho UI — token màu/spacing/typography cụ thể xem [design-system.md](design-system.md).
- **Zod schema** làm hợp đồng validation dùng chung giữa FE/BE cho các form nhập liệu (tạo sự kiện, checkout...) — schema Zod nên phản ánh đúng schema request body định nghĩa trong [../../api-docs/openapi/](../../api-docs/openapi/), không tự định nghĩa validation khác biệt.
- Quy tắc hành vi tương tác dùng chung (loading, toast, modal, validate form, pagination...) xem [interaction-patterns.md](interaction-patterns.md).
- Icon: `lucide-react` — bộ icon duy nhất dùng trong toàn bộ frontend, xem [design-system.md](design-system.md#iconography).

## Auth ở phía frontend

- **Access token**: giữ ở bộ nhớ (in-memory, vd: trong React context/store), **không** lưu ở `localStorage`/`sessionStorage` để giảm rủi ro bị đánh cắp qua XSS.
- **Refresh token**: cookie `httpOnly` do backend set — frontend không đọc/ghi trực tiếp cookie này; chỉ gọi `POST /auth/refresh` và để browser tự đính kèm cookie.
- Khi access token hết hạn (401 từ API), tự động gọi `/auth/refresh` một lần để lấy access token mới rồi retry request gốc; nếu refresh cũng thất bại, chuyển hướng người dùng về [screens/public/auth-login.md](screens/public/auth-login.md).

Chi tiết luồng auth đầy đủ xem [../04-security/authentication.md](../04-security/authentication.md).

## CORS & ảnh

- CORS: API Gateway chỉ định đúng domain frontend được phép gọi kèm `credentials: true` (bắt buộc vì cookie refresh token là cross-origin nếu frontend và API khác domain/subdomain).
- Ảnh (banner sự kiện, avatar) phục vụ qua Next.js Image Optimization kết hợp Cloudflare CDN — không tự implement resize ảnh ở backend, tận dụng object storage + CDN có sẵn (xem [../02-domains/file-storage/spec.md](../02-domains/file-storage/spec.md)) (khi đã có banner/avatar — Phase 2; Phase 1 chưa có trường nhập ảnh).

## Mục lục UI spec

| Thư mục | Nội dung |
|---|---|
| [design-system.md](design-system.md) | Token màu (theo thang Tailwind), typography, spacing, radius/shadow, breakpoint, icon set |
| [interaction-patterns.md](interaction-patterns.md) | Quy tắc hành vi dùng chung: loading, toast, modal/dialog, confirm dialog, validate form, empty/error state, pagination, optimistic update |
| [components/](components/) | Component dùng chung giữa nhiều màn hình (Header, Footer, Button, Input, Card, Modal, Toast...) — **đọc trước khi viết bất kỳ screen spec nào** |
| [screens/](screens/) | Đặc tả từng màn hình cụ thể, có bảng chỉ mục và quy trình thêm màn hình mới |

Khi thêm màn hình hoặc component UI mới, làm theo quy trình đã mô tả tại [screens/README.md](screens/README.md#quy-trình-thêm-màn-hình-mới) và [components/README.md](components/README.md#quy-trình-thêm-component-mới) — xem thêm mục "Thêm màn hình / component UI mới" tại [../../AGENTS.md](../../AGENTS.md).
