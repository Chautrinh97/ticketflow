---
name: ui-spec
description: Write, update, or audit a docs/06-frontend/screens/*.md or docs/06-frontend/components/*.md file against the mandatory screen/component templates. Use when documenting a new or changed screen/component, or reviewing existing UI docs for template drift, duplicated components, or mismatch against the implemented Next.js code.
---

# ui-spec

Viết hoặc audit file dưới `docs/06-frontend/screens/` hoặc `docs/06-frontend/components/`. **Luôn đọc trước** [screens/README.md](../../../docs/06-frontend/screens/README.md) (khuôn mẫu screen + quy trình) hoặc [components/README.md](../../../docs/06-frontend/components/README.md) (khuôn mẫu component + quy trình) tuỳ mục tiêu — hai file đó đã có đầy đủ khuôn mẫu, skill này không hardcode lại.

## Chế độ: viết mới / cập nhật

1. **Rà soát bắt buộc trước tiên**: đọc cả 4 file `docs/06-frontend/components/{layout,buttons-inputs,feedback,data-display}.md` trước khi mô tả bất kỳ thành phần giao diện nào (nguyên tắc quan trọng nhất trong AGENTS.md, lặp lại nguyên văn ở CLAUDE.md). Có component tương đương → chỉ tham chiếu tên + variant, không mô tả lại.
2. Chỉ tách component dùng chung mới khi **thực sự có ≥2 màn hình dự kiến dùng** (`components/README.md` bước 3) — nếu chỉ 1 màn hình dùng, mô tả trực tiếp trong "Chi tiết từng section riêng" của chính screen spec đó.
3. Viết screen spec theo đúng 6 phần cố định (Header → Layout tổng quan → Chi tiết từng section riêng → Hành vi tương tác → Trạng thái đặc biệt → Responsive) theo `screens/README.md`. Viết component doc theo đúng khuôn 5 phần (mục đích — danh sách màn hình dùng — variant/size/state — cấu trúc trực quan — hành vi riêng) theo `components/README.md`.
4. Màu/spacing/typography chỉ lấy token từ `docs/06-frontend/design-system.md`; hành vi loading/toast/modal/validate/pagination chỉ tham chiếu `docs/06-frontend/interaction-patterns.md`, không tự định nghĩa lại quy tắc chung trong screen spec.
5. Nếu cần API chưa tồn tại: dừng lại, dùng skill `api-spec` để cập nhật OpenAPI trước — không viết UI dựa trên hợp đồng chưa có (`screens/README.md` bước 3).
6. Cập nhật đúng bảng chỉ mục (`screens/README.md` hoặc `components/README.md`'s "Danh sách file") trong cùng thay đổi.

## Chế độ: audit

- Kiểm tra đủ 6 phần (screen) / 5 phần (component), đúng thứ tự, đúng heading.
- **Trùng lặp component**: rà `src/frontend/nextjs-app/components/` thật — flag khi 1 screen spec tự mô tả chi tiết 1 phần tử mà thực chất trùng cấu trúc/hành vi với component đã có trong `components/*.md` (lẽ ra phải tham chiếu, không mô tả lại).
- **Component chỉ dùng 1 nơi**: flag component doc có "danh sách màn hình dùng" chỉ liệt kê đúng 1 màn hình — vi phạm `components/README.md` bước 3.
- **Orphan 2 chiều**: đối chiếu bảng chỉ mục với file thật trong `screens/`/`components/` — flag file tồn tại nhưng thiếu dòng trong bảng, và dòng trong bảng nhưng file không tồn tại.
- **Khớp domain spec + OpenAPI + route**: kiểm tra link trong header mỗi screen spec (domain spec, OpenAPI, route) vẫn còn đúng và tồn tại thật.
- **Khớp code thật**: khi màn hình đã implement (`src/frontend/nextjs-app/app/`), đối chiếu "Trạng thái đặc biệt" (loading/empty/error) và "Hành vi tương tác" mô tả trong spec với behavior thật trong code — flag phần mô tả trong spec nhưng chưa implement, hoặc implement nhưng spec không nhắc tới.
- Báo cáo drift theo dạng: `[screens|components]/<file> — spec nói X, thực tế Y`.
