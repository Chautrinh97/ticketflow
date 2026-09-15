---
name: ui-dev
description: Implement or modify a Next.js screen or component under src/frontend/nextjs-app/, following docs/06-frontend conventions — component reuse, design tokens, interaction patterns, Zod/OpenAPI alignment, and the apiFetch/token conventions. Use when writing new frontend UI code.
---

# ui-dev

Implement code dưới `src/frontend/nextjs-app/`. Đọc [docs/06-frontend/README.md](../../../docs/06-frontend/README.md) (đặc biệt mục "Quy ước code") trước khi bắt đầu.

## Checklist bắt buộc

1. **Rà soát component trước tiên**: đọc cả 4 file `docs/06-frontend/components/{layout,buttons-inputs,feedback,data-display}.md` trước khi implement bất kỳ phần UI nào. Có sẵn Header/Button/Input/Card... → chỉ dùng lại (đúng tên + variant/props), không tự viết component khác làm việc tương tự (nguyên tắc quan trọng nhất trong AGENTS.md).
2. Nếu đang implement 1 screen mới: đọc trước `docs/06-frontend/screens/<role>/<screen>.md` tương ứng (nếu chưa có, dùng skill `ui-spec` để viết trước) — code phải khớp đúng 6 phần đã đặc tả (layout, section, hành vi tương tác, trạng thái đặc biệt, responsive).
3. **Token & hành vi dùng chung**: màu/spacing/typography chỉ lấy từ `docs/06-frontend/design-system.md` (không tự chọn giá trị Tailwind tuỳ ý); hành vi loading/toast/modal/validate/pagination/optimistic update theo đúng `docs/06-frontend/interaction-patterns.md`, không tự sáng tạo pattern riêng cho 1 màn hình.
4. **Gọi API**: mọi lời gọi API đi qua `apiFetch<T>` trong `lib/api/client.ts` (không `fetch` thô); hàm gọi API mới đặt trong `lib/api/<domain>.ts`, dùng helper `toQuery` cho query string, tham số list dùng đúng tên `page`/`page_size`.
5. **Schema**: `lib/schemas/<domain>.schema.ts` (Zod) phải khớp field-for-field với request/response schema trong `api-docs/openapi/<service>.yaml` tương ứng — field mới bên nào thì cập nhật bên kia trong cùng thay đổi.
6. **Data fetching**: cache/refetch phía client qua hook riêng trong `lib/hooks/` bọc TanStack Query, không tự gọi `apiFetch` trực tiếp rải rác trong component khi có thể tái sử dụng qua hook.
7. **Auth**: access token chỉ đọc/ghi qua `tokenStore` (in-memory) — **không** lưu vào `localStorage`/`sessionStorage`; không tự đọc/ghi cookie refresh token (đó là việc của `apiFetch`/backend).
8. **Icon**: chỉ dùng `lucide-react`, không thêm bộ icon khác.
9. Nếu cần API endpoint chưa tồn tại: dừng lại, dùng skill `api-spec` cập nhật OpenAPI trước — không build UI dựa trên hợp đồng chưa định nghĩa.
10. Nếu thêm screen/component mới hoặc đổi hành vi đáng kể: cập nhật/viết doc tương ứng bằng skill `ui-spec` trong cùng thay đổi.
