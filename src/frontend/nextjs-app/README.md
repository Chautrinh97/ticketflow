# Next.js App

**Trạng thái:** Phase 1 (MVP) — 12 màn hình đã implement: đăng nhập (mock Firebase mặc định, xem `lib/firebase.ts`), trang chủ, tìm kiếm sự kiện (lọc cơ bản), chi tiết sự kiện, đặt vé (nhiều loại vé/1 đơn — xem `app/events/[slug]/checkout`), xác nhận đơn hàng, vé của tôi, chi tiết vé (QR code), hồ sơ, tạo sự kiện, quản lý sự kiện cơ bản, trạng thái lỗi (404/runtime). Chạy: `npm install && npm run dev` (hoặc qua `deployments/docker-compose.yaml`). Chưa implement: dashboard organizer/admin đầy đủ, tìm kiếm full-text, thông báo, huỷ vé (Phase 2).

Frontend TicketFlow — Next.js 14 (App Router) + Tailwind CSS + TanStack Query.

- Tổng quan kỹ thuật (render mode theo trang, auth, CORS, ảnh): [../../../docs/06-frontend/README.md](../../../docs/06-frontend/README.md)
- Đặc tả từng màn hình UI: [../../../docs/06-frontend/screens/](../../../docs/06-frontend/screens/)
- Component dùng chung (Header, Button, Input, Card...): [../../../docs/06-frontend/components/](../../../docs/06-frontend/components/)
- Design system (màu/typography/spacing): [../../../docs/06-frontend/design-system.md](../../../docs/06-frontend/design-system.md)
- Hợp đồng API dùng để gọi backend: [../../../api-docs/](../../../api-docs/)
