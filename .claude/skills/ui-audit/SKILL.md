---
name: ui-audit
description: Audit existing Next.js frontend code for component-reuse violations, design-token drift, unsafe token storage, raw fetch calls, and Zod/OpenAPI schema mismatches. Use for reviewing frontend UI code changes or an area of the app, without writing new features.
---

# ui-audit

Audit code dưới `src/frontend/nextjs-app/`. Đọc [docs/06-frontend/README.md](../../../docs/06-frontend/README.md) trước để biết đúng quy ước cần đối chiếu.

## Giới hạn cần nêu rõ khi báo cáo

Repo chỉ có `next lint` (ESLint, `next/core-web-vitals`) và `tsc --noEmit` — **không có Prettier, không có test framework nào** (Jest/Vitest/Playwright/Testing Library đều chưa cài). Skill này không thể kiểm tra "test có cover hành vi mới không" vì chưa có test nào để đối chiếu; mọi check dưới đây làm bằng đọc/grep thủ công. Có thể gợi ý chạy `npm run lint` và `npm run typecheck` như bước bổ trợ, nhưng đó không thay thế được checklist dưới đây.

## Checklist audit

- **Trùng lặp component**: grep `components/`/`app/` tìm phần tử tự viết (vd `<button className=...>` thô) trong khi đã có component tương đương (`Button`, `Input`...) mô tả ở `docs/06-frontend/components/*.md` — flag vi phạm nguyên tắc tái sử dụng của AGENTS.md.
- **Lệch Zod/OpenAPI**: đối chiếu field trong từng `lib/schemas/<domain>.schema.ts` với request/response schema tương ứng trong `api-docs/openapi/<service>.yaml` — flag field thiếu/thừa/sai kiểu/sai tên (snake_case).
- **Token lưu sai chỗ**: grep `localStorage`/`sessionStorage` tìm việc ghi access/refresh token — vi phạm quy tắc auth trong `docs/06-frontend/README.md`.
- **Gọi API không qua wrapper**: grep các lời gọi `fetch(` ngoài `lib/api/client.ts` — flag nơi gọi `fetch` thô thay vì qua `apiFetch`.
- **Token design-system**: grep giá trị Tailwind tuỳ ý bất thường (màu hex trực tiếp, spacing lẻ không theo thang chuẩn) thay vì token đã định nghĩa ở `design-system.md`.
- **Đối chiếu spec UI**: cross-check với `docs/06-frontend/screens/<role>/<screen>.md` tương ứng (nếu có) — flag hành vi/trạng thái đặc biệt mô tả trong spec nhưng chưa implement, hoặc implement nhưng spec chưa mô tả (dùng cùng lúc với skill `ui-spec` chế độ audit để có coverage đủ 2 chiều).
- Báo cáo theo dạng: `<file>:<dòng nếu có> — <vi phạm cụ thể>, tham chiếu quy ước nào`.
