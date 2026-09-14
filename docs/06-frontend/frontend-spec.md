# Frontend (Next.js)

## Render mode theo trang

| Trang | Render mode | Lý do |
|---|---|---|
| Trang chủ / danh sách sự kiện | ISR (Incremental Static Regeneration) | Cập nhật định kỳ, tốt cho SEO, không cần dữ liệu real-time tuyệt đối |
| Chi tiết sự kiện | SSR | Cần dữ liệu mới nhất (tồn kho vé còn lại) tại thời điểm request |
| Trang đặt vé/checkout | CSR | Tương tác nhiều bước, cần auth, không cần SEO |
| Vé của tôi | CSR (cần auth) | Dữ liệu cá nhân hoá, không cần SEO |
| Dashboard organizer/admin | CSR (cần auth) | Dữ liệu cá nhân hoá, tương tác nhiều |

## Chi tiết kỹ thuật

- **Next.js App Router**, dùng React Server Components cho các trang cần SEO (trang chủ, chi tiết sự kiện) để giảm JS gửi xuống client.
- **TanStack Query (React Query)** quản lý cache dữ liệu phía client cho các trang CSR — đặc biệt cần cho trạng thái tồn kho vé (`ticket_types.sold_count`/`quota`) có thể thay đổi trong lúc người dùng đang ở trang checkout.
- **Tailwind CSS** cho UI.
- **Zod schema** làm hợp đồng validation dùng chung giữa FE/BE cho các form nhập liệu (tạo sự kiện, checkout...) — schema Zod nên phản ánh đúng schema request body định nghĩa trong [../../api-docs/openapi/](../../api-docs/openapi/), không tự định nghĩa validation khác biệt.

## Auth ở phía frontend

- **Access token**: giữ ở bộ nhớ (in-memory, vd: trong React context/store), **không** lưu ở `localStorage`/`sessionStorage` để giảm rủi ro bị đánh cắp qua XSS.
- **Refresh token**: cookie `httpOnly` do backend set — frontend không đọc/ghi trực tiếp cookie này; chỉ gọi `POST /auth/refresh` và để browser tự đính kèm cookie.
- Khi access token hết hạn (401 từ API), tự động gọi `/auth/refresh` một lần để lấy access token mới rồi retry request gốc; nếu refresh cũng thất bại, chuyển hướng người dùng về trang đăng nhập.

Chi tiết luồng auth đầy đủ xem [../04-security/authentication.md](../04-security/authentication.md).

## CORS & ảnh

- CORS: API Gateway chỉ định đúng domain frontend được phép gọi kèm `credentials: true` (bắt buộc vì cookie refresh token là cross-origin nếu frontend và API khác domain/subdomain).
- Ảnh (banner sự kiện, avatar) phục vụ qua Next.js Image Optimization kết hợp Cloudflare CDN — không tự implement resize ảnh ở backend, tận dụng object storage + CDN có sẵn (xem [../02-domains/file-storage/spec.md](../02-domains/file-storage/spec.md)).
