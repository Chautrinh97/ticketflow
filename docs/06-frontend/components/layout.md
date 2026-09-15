# Layout components

Token dùng: [../design-system.md](../design-system.md). Quy tắc chung: [../interaction-patterns.md](../interaction-patterns.md).

## Header

**Mục đích:** thanh điều hướng trên cùng, sticky khi cuộn (`sticky top-0 z-40 bg-white border-b border-gray-200`, thêm `shadow-sm` sau khi cuộn quá 80px).

**Dùng ở:** mọi màn hình, theo 2 variant:

| Variant | Dùng ở |
|---|---|
| `public` | Toàn bộ `screens/public/*`, `screens/user/*` |
| `dashboard` | Toàn bộ `screens/organizer/*`, `screens/admin/*` (đi kèm `Sidebar`) |

**Cấu trúc — variant `public`:**

| Vị trí | Thành phần |
|---|---|
| Trái | Logo TicketFlow (link `/`) |
| Giữa | Nav link chính (vd "Sự kiện") — ẩn dưới `md`, gộp vào menu hamburger |
| Phải | `SearchBar` (thu gọn, xem [buttons-inputs.md](buttons-inputs.md)) · `NotificationBell` (chỉ khi đã đăng nhập) · `UserMenu`/nút "Đăng nhập" |

**Cấu trúc — variant `dashboard`:**

| Vị trí | Thành phần |
|---|---|
| Trái | Nút toggle `Sidebar` (chỉ hiện dưới `lg`) · `Breadcrumb` |
| Phải | `NotificationBell` · `UserMenu` (rút gọn: "Về trang chính", "Đăng xuất") |

**`NotificationBell`** (Phase 2 — cần Notification Service): icon chuông (lucide `bell`) + badge số thông báo chưa đọc (nền `danger`, ẩn nếu = 0). Click mở dropdown panel: 5 thông báo gần nhất (rút gọn từ [../screens/user/notifications.md](../screens/user/notifications.md)) + link "Xem tất cả" điều hướng sang trang đó.

**`UserMenu`:** avatar (`Avatar`, xem [data-display.md](data-display.md)) + tên. Click mở dropdown:
- Luôn có: "Hồ sơ của tôi" → profile.md, "Vé của tôi" → my-bookings.md, "Đăng xuất".
- Nếu `role=user`: thêm "Trở thành Organizer" → organizer-request.md (Phase 2).
- Nếu `role=organizer`: thêm "Kênh Organizer" → event-form.md (tạo sự kiện, Phase 1); từ Phase 2 trỏ sang organizer/dashboard.md đầy đủ.
- Nếu `role=super_admin`: thêm "Quản trị hệ thống" → admin/dashboard.md (Phase 2).

Chưa đăng nhập: thay `NotificationBell` + `UserMenu` bằng nút "Đăng nhập" (`Button` variant `primary`, size `sm`) → auth-login.md.

## Footer

**Mục đích:** chân trang. **Chỉ** dùng cùng Header variant `public` — **không** hiển thị ở layout dashboard (nhường không gian cho nội dung quản trị).

**Cấu trúc:** 3 cột (giới thiệu ngắn về TicketFlow · liên kết nhanh: Về chúng tôi/Điều khoản/Liên hệ · mạng xã hội), dưới `md` xếp dọc 1 cột. Dòng cuối: copyright, căn giữa, `text-xs text-gray-400`.

Không có hành vi tương tác đặc biệt ngoài điều hướng link thường.

## Sidebar

**Mục đích:** điều hướng chính trong khu vực quản trị, đi kèm Header variant `dashboard` (Phase 2 — dashboard organizer/admin chưa có ở Phase 1).

| Variant | Dùng ở | Menu item |
|---|---|---|
| `organizer` | `screens/organizer/*` | Tổng quan · Sự kiện của tôi |
| `admin` | `screens/admin/*` | Tổng quan · Người dùng · Audit log |

**Cấu trúc:** cố định bên trái, rộng `w-60` (240px) từ `lg` trở lên; dưới `lg` ẩn mặc định, mở dưới dạng drawer trượt từ trái (overlay token `overlay`) khi bấm nút toggle trên Header. Item đang active: nền `bg-blue-50`, chữ + icon `text-blue-600` (token `primary`); item thường: `text-gray-700`, hover `bg-gray-50`.

**Hành vi:** click item → điều hướng route tương ứng; nếu đang ở chế độ drawer (mobile), tự đóng drawer sau khi điều hướng.

## Breadcrumb

**Mục đích:** định vị vị trí hiện tại trong dashboard, vd `Sự kiện của tôi / Đêm Nhạc Trẻ 2026 / Chỉnh sửa`.

**Dùng ở:** mọi `screens/organizer/*`, `screens/admin/*` (trong Header variant `dashboard`) (Phase 2 — dashboard organizer/admin chưa có ở Phase 1).

**Cấu trúc:** chuỗi `text-sm`, các mục không phải cuối là link `text-gray-500 hover:text-gray-700`, dấu phân cách `/` (`text-gray-300`), mục cuối (trang hiện tại) `text-gray-900 font-medium`, không phải link.

## PageContainer

**Mục đích:** wrapper căn giữa nội dung, giới hạn max-width, padding ngang nhất quán.

| Variant | Class |
|---|---|
| `public` (trang public/user) | `max-w-6xl mx-auto px-4` |
| `dashboard` (trang organizer/admin, đã có Sidebar giới hạn không gian) | `max-w-full px-6` |

Không có hành vi tương tác.
