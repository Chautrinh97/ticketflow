# Design System

Bộ token thị giác dùng chung cho toàn bộ frontend TicketFlow — mọi màn hình ([screens/](screens/)) và component dùng chung ([components/](components/)) phải dùng token tại đây, không tự chọn giá trị màu/spacing riêng. Token bám theo thang mặc định của **Tailwind CSS** (đã chọn ở [README.md](README.md)) để implement được ngay, không cần custom `tailwind.config`.

## Màu sắc

| Token semantic | Dùng cho | Class Tailwind | Ghi chú trạng thái |
|---|---|---|---|
| `primary` | CTA chính (nút Đặt vé, Lưu, Xuất bản...), link nhấn mạnh, mục nav đang active | `bg-blue-600` / `text-blue-600` | hover `blue-700`, active/pressed `blue-800` |
| `primary-foreground` | Chữ/icon đặt trên nền `primary` | `text-white` | |
| `secondary` | CTA phụ (nút Huỷ, outline button) | `border-gray-300 text-gray-700` | hover `bg-gray-50` |
| `danger` | Hành động phá huỷ (xoá/huỷ sự kiện, huỷ vé, khoá tài khoản), thông báo lỗi | `bg-red-600` / `text-red-600` | hover `red-700` |
| `success` | Trạng thái thành công (`paid`, `published`, `valid`, `active`) | `bg-green-600` / `text-green-600` | |
| `warning` | Trạng thái cần chú ý (`pending`, sắp hết vé, sắp hết hạn thanh toán) | `bg-amber-500` / `text-amber-600` | |
| `text-primary` | Chữ nội dung chính | `text-gray-900` | |
| `text-secondary` | Chữ phụ, meta info, mô tả ngắn | `text-gray-500` | |
| `text-disabled` | Chữ ở trạng thái disabled | `text-gray-400` | |
| `bg-base` | Nền trang mặc định | `bg-white` | |
| `bg-subtle` | Nền section/card phụ, phân tách khối nội dung | `bg-gray-50` | |
| `border` | Viền input/card/divider | `border-gray-200` | |
| `focus-ring` | Viền focus khi điều hướng bàn phím (input, button, link) | `ring-2 ring-blue-500 ring-offset-2` | bắt buộc cho mọi phần tử tương tác được, không được bỏ để giữ accessibility |
| `overlay` | Nền mờ phía sau Modal/Dialog | `bg-black/50` | |

Trạng thái (`draft`/`published`/`cancelled`, `pending`/`paid`/`expired`, `valid`/`used`...) không có màu riêng — luôn map về 1 trong 4 token `success`/`warning`/`danger`/neutral (`text-secondary` + `bg-subtle`). Bảng map cụ thể từng trạng thái nằm ở component `Badge/StatusTag` — xem [components/data-display.md](components/data-display.md).

## Typography

| Style | Dùng cho | Class Tailwind |
|---|---|---|
| Display | Tiêu đề hero (banner trang chủ) | `text-4xl font-bold tracking-tight` |
| H1 | Tiêu đề trang | `text-2xl font-bold` |
| H2 | Tiêu đề section | `text-xl font-semibold` |
| H3 | Tiêu đề panel/card | `text-lg font-semibold` |
| Body | Nội dung chính | `text-base font-normal` |
| Body small | Mô tả phụ, meta info | `text-sm text-gray-500` |
| Caption | Label nhỏ, timestamp, helper text dưới input | `text-xs text-gray-400` |
| Button text | Chữ trong Button | `text-sm font-medium` |

## Spacing & layout grid

Dùng trực tiếp thang spacing mặc định của Tailwind (bội số 4px), không định nghĩa thang riêng. Quy ước áp dụng nhất quán:

| Ngữ cảnh | Class gợi ý |
|---|---|
| Khoảng cách giữa các section lớn trong 1 trang | `space-y-8` / `gap-8` (32px) |
| Khoảng cách giữa các item trong 1 grid/list (event card, ticket type row) | `gap-4` đến `gap-6` (16–24px) |
| Padding trong card/panel | `p-4` (compact, bảng dữ liệu) hoặc `p-6` (card nổi bật) |
| Padding trong Button/Input | size `md`: `px-4 py-2` · size `sm`: `px-3 py-1.5` |
| Max width nội dung chính (container) | `max-w-6xl mx-auto px-4` |

## Border radius & shadow (elevation)

| Cấp độ | Radius | Shadow | Dùng cho |
|---|---|---|---|
| Nhỏ | `rounded-md` (6px) | — | Input, Button, Badge |
| Vừa | `rounded-lg` (8px) | `shadow-sm` mặc định, `shadow-md` khi hover | Card, Panel |
| Lớn | `rounded-xl` (12px) | `shadow-xl` | Modal, Popover, Dropdown menu |

## Breakpoints

Theo mặc định Tailwind: `sm` 640px · `md` 768px · `lg` 1024px · `xl` 1280px. Thiết kế **mobile-first**: layout mặc định 1 cột, chuyển sang nhiều cột từ `md` trở lên (vd danh sách sự kiện: 1 cột dưới `md`, 2 cột `md`, 3 cột `lg`). Sidebar dashboard (organizer/admin) hiển thị cố định từ `lg` trở lên, thu gọn thành drawer trượt (off-canvas) dưới `lg` — chi tiết ở [components/layout.md](components/layout.md).

## Iconography

Dùng **một bộ icon duy nhất**: [`lucide-react`](https://lucide.dev) — phổ biến với stack Tailwind/Next.js, tránh trộn nhiều icon set gây lệch độ dày nét/kích thước. Kích thước chuẩn:

| Ngữ cảnh | Kích thước |
|---|---|
| Icon trong Button/Input/nav item | `w-5 h-5` (20px) |
| Icon nhỏ trong Badge/caption/meta | `w-4 h-4` (16px) |
| Icon trong Empty state | `w-6 h-6` (24px) |
