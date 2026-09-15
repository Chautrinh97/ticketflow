# Button & Input components

Token dùng: [../design-system.md](../design-system.md). Quy tắc chung (validate form, loading): [../interaction-patterns.md](../interaction-patterns.md).

## Button

**Mục đích:** hành động chính/phụ trên mọi màn hình — CTA "Đặt vé ngay", "Lưu", "Xuất bản", nút trong `ConfirmDialog`, v.v. **Dùng ở:** hầu hết mọi screen spec.

| Variant | Dùng khi | Class nền/chữ |
|---|---|---|
| `primary` | Hành động chính của màn hình/form (tối đa 1 nút `primary` mỗi khung nhìn) | `bg-blue-600 text-white`, hover `blue-700`, active `blue-800` |
| `secondary` | Hành động phụ (Huỷ, Quay lại) | `border border-gray-300 text-gray-700`, hover `bg-gray-50` |
| `danger` | Hành động phá huỷ (xác nhận xoá/huỷ/khoá trong `ConfirmDialog`) | `bg-red-600 text-white`, hover `red-700` |
| `ghost` | Hành động phụ nhẹ, không viền (vd "Bỏ qua") | `text-gray-700`, hover `bg-gray-100` |

Size: `sm` (`px-3 py-1.5 text-sm`) · `md` (`px-4 py-2 text-sm`, mặc định) · `lg` (`px-6 py-3 text-base`).

**Trạng thái:** `disabled` (`opacity-50 pointer-events-none`) · `loading` (icon trái thay bằng spinner quay, giữ nguyên text, button tự chuyển `disabled`) — không thay đổi chiều rộng button khi chuyển sang loading, tránh giật layout.

## IconButton

**Mục đích:** hành động phụ chỉ cần icon — sửa/xoá trong dòng bảng (`DataTable`), đóng `Modal` (X), toggle `Sidebar` mobile. **Dùng ở:** `data-display.md` → `DataTable`, `feedback.md` → `Modal`, `layout.md` → `Header`/`Sidebar`.

Size: `sm` (32px) · `md` (40px, mặc định). Variant: `ghost` (mặc định, nền trong suốt, hover `bg-gray-100`) · `danger-ghost` (icon `text-red-600`, dùng cho nút xoá trong bảng).

## Input (text / email / password / textarea / number)

**Mục đích:** nhập liệu văn bản trong mọi form. **Dùng ở:** auth-login.md (email), profile.md, event-form.md, checkout.md (nếu cần ghi chú), mọi form khác.

**Cấu trúc:** label phía trên (`text-sm font-medium text-gray-700`) → ô nhập (`border border-gray-200 rounded-md px-4 py-2`) → helper/caption text dưới (`text-xs text-gray-400`).

**Trạng thái:** `default` · `focus` (áp token `focus-ring`) · `disabled` (`bg-gray-50 text-gray-400`) · `error` (border `border-red-500`, helper text thay bằng `InlineFormError` — xem [feedback.md](feedback.md)) · `readonly`.

**Variant riêng:**
- `password`: icon con mắt bên phải để toggle hiện/ẩn.
- `textarea`: nhiều dòng, resize dọc, dùng cho mô tả sự kiện.
- `number`: dùng cho các trường số không phải số lượng vé (giá vé, %...) — số lượng vé dùng `QuantityStepper` riêng bên dưới, không dùng `Input number` thường.

## QuantityStepper

**Mục đích:** chọn số lượng vé. **Dùng ở:** checkout.md.

**Cấu trúc:** `IconButton ghost` (icon `minus`) — số hiển thị giữa (`text-lg font-semibold`, có thể nhập tay) — `IconButton ghost` (icon `plus`).

**Hành vi:** `min=1`; `max = min(10, số vé loại đó còn lại)` — nút trừ disable khi chạm `min`, nút cộng disable khi chạm `max`; nhập tay giá trị ngoài khoảng thì tự **clamp** về giới hạn hợp lệ khi rời khỏi ô (blur).

## Select / Dropdown

**Mục đích:** chọn 1 giá trị trong danh sách cố định — filter (category/city/status), chọn role/status trong form admin. **Dùng ở:** event-search.md, event-form.md, user-management.md (Phase 2), my-bookings.md (filter status).

**Cấu trúc:** giống `Input` nhưng có icon chevron-down bên phải; click mở menu dropdown (`rounded-lg shadow-xl`, elevation lớn theo design-system), mỗi option cao `py-2 px-4`, hover `bg-gray-50`, option đang chọn có dấu check bên phải.

## DatePicker

**Mục đích:** chọn ngày/khoảng ngày. **Dùng ở:** event-form.md (thời gian bắt đầu/kết thúc sự kiện), event-search.md (lọc theo ngày).

**Cấu trúc:** input dạng text hiển thị ngày đã chọn (định dạng `dd/MM/yyyy`) + icon lịch bên phải; click mở calendar popover (`rounded-lg shadow-xl`).

## SearchBar

**Mục đích:** tìm kiếm sự kiện theo từ khoá. **Dùng ở:** `Header` (variant `public`, dạng thu gọn) và đầu trang event-search.md (dạng đầy đủ, có thêm nút "Tìm kiếm").

**Cấu trúc:** input bo tròn hết cỡ (`rounded-full`), icon kính lúp bên trái, placeholder "Tìm sự kiện, nghệ sĩ, địa điểm...".

**Hành vi:** nhấn `Enter` hoặc click icon kính lúp → điều hướng sang event-search.md kèm query `?q=`.

## FileUpload / ImageUploader

**Mục đích:** chọn & upload ảnh (banner sự kiện, avatar). **Dùng ở:** event-form.md (banner), profile.md (avatar) (cả hai: Phase 2).

**Cấu trúc:** khung tỉ lệ 16:9 (banner) hoặc tròn (avatar), viền `border-2 border-dashed border-gray-300`, click hoặc kéo-thả file vào để chọn. Sau khi chọn: hiện preview ảnh + progress bar khi đang upload + `IconButton` đổi ảnh/xoá đè góc trên phải.

**Hành vi:** validate loại file (`image/jpeg`, `image/png`, `image/webp`) và kích thước ngay phía client (khớp whitelist tại [../../02-domains/file-storage/spec.md](../../02-domains/file-storage/spec.md)) trước khi gọi `POST /files/presign`; nếu sai, hiện `InlineFormError` ngay dưới khung, không gọi API. Sau khi có `upload_url`, `PUT` file trực tiếp lên đó rồi lưu `public_url` vào state form — không upload qua backend.

## Checkbox / Radio

**Mục đích:** chọn nhiều/chọn 1 trong nhóm nhỏ lựa chọn — filter đa lựa chọn, chọn phương án trong form. **Dùng ở:** event-search.md (filter category nếu cho chọn nhiều).

**Cấu trúc:** theo chuẩn form mặc định của Tailwind (`accent-blue-600`), áp token `focus-ring` khi focus bằng bàn phím.
