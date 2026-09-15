# Data display components

Token dùng: [../design-system.md](../design-system.md). Quy tắc pagination: [../interaction-patterns.md](../interaction-patterns.md#pagination).

## EventCard

**Mục đích:** thẻ tóm tắt 1 sự kiện trong danh sách/lưới. **Dùng ở:** home.md, event-search.md.

**Cấu trúc:** khối `rounded-lg shadow-sm hover:shadow-md border border-gray-200`, ảnh banner tỉ lệ 16:9 ở trên (bo góc trên) → phần nội dung `p-4`: `Badge` category (góc trên ảnh, đè lên ảnh) → tiêu đề sự kiện (`H3`, tối đa 2 dòng, cắt `line-clamp-2`) → dòng địa điểm + ngày (`Body small`, icon `map-pin`/`calendar`) → giá vé thấp nhất (`text-primary font-semibold`, định dạng "Từ 200.000đ").

**Hành vi:** cả thẻ là 1 link tới event-detail.md; hover nâng `shadow-md` nhẹ (`transition-shadow`).

## TicketTypeRow

**Mục đích:** 1 dòng thể hiện 1 loại vé (tên, giá, số lượng còn lại) kèm chọn số lượng. **Dùng ở:** event-detail.md (chỉ xem), checkout.md (chọn số lượng qua `QuantityStepper`, xem [buttons-inputs.md](buttons-inputs.md)), event-manage.md (organizer xem/thêm ở Phase 1; sửa nhanh ở Phase 2).

**Cấu trúc:** hàng ngang `border-b border-gray-200 py-4`, trái: tên loại vé (`H3`) + trạng thái còn/hết (`Body small`, `warning` nếu còn ≤ 10% quota, `danger` "Hết vé" nếu `sold_count == quota`); phải: giá (`font-semibold`) + `QuantityStepper` (chỉ ở checkout.md) hoặc số đã bán/tổng (chỉ ở event-manage.md, dạng `sold_count/quota`).

## Badge / StatusTag

**Mục đích:** nhãn trạng thái ngắn, màu theo ý nghĩa. **Dùng ở:** EventCard, event-manage.md, my-bookings.md, booking-detail.md, user-management.md (Phase 2), mọi nơi hiển thị trạng thái.

**Cấu trúc:** `inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium`, nền nhạt + chữ đậm cùng tông (vd `success` → `bg-green-50 text-green-700`).

**Bảng map trạng thái nghiệp vụ → token màu** (nguồn giá trị enum: xem `docs/03-data/postgres-schema.md`):

| Nhóm | Giá trị | Token |
|---|---|---|
| Sự kiện (`events.status`) | `draft` | neutral (`bg-gray-100 text-gray-600`) |
| | `published` | `success` |
| | `cancelled` | `danger` |
| Đơn hàng (`orders.status`) | `pending` | `warning` |
| | `paid` | `success` |
| | `cancelled` / `expired` | `danger` |
| Vé (`tickets.status`) | `valid` | `success` |
| | `used` | neutral |
| | `cancelled` | `danger` |
| Tài khoản (`users.status`) | `active` | `success` |
| | `pending` | `warning` |
| | `banned` | `danger` |

## DataTable

**Mục đích:** bảng dữ liệu cho khu vực quản trị (nhiều cột, sort, phân trang theo số trang). **Dùng ở:** event-list.md, event-manage.md (tab Người mua vé) (Phase 2), user-management.md (Phase 2), audit-log.md (Phase 2).

**Cấu trúc:** header cột `bg-gray-50 text-xs font-medium text-gray-500 uppercase`, mỗi dòng `border-b border-gray-100 hover:bg-gray-50`, cột hành động cuối cùng chứa `IconButton` (sửa/xoá/xem chi tiết). Cột có thể sort: icon mũi tên nhỏ cạnh tên cột, click để đổi chiều sort.

**Hành vi:** click vào 1 dòng (trừ cột hành động) → điều hướng sang trang chi tiết tương ứng nếu có; rỗng → `EmptyState` (xem [feedback.md](feedback.md)); đang tải → `Skeleton` dạng nhiều dòng.

## Pagination

**Mục đích:** điều hướng trang cho `DataTable` (quy ước phân trang theo số trang — xem [../interaction-patterns.md](../interaction-patterns.md#pagination)). **Dùng ở:** cùng danh sách với `DataTable`.

**Cấu trúc:** căn phải dưới bảng — nút "‹ Trước", danh sách số trang (rút gọn bằng `...` nếu nhiều), nút "Sau ›"; trang hiện tại `bg-blue-600 text-white rounded-md`.

Với danh sách dùng kiểu **"Xem thêm"** (home.md, event-search.md, my-bookings.md, notifications.md): dùng `Button` variant `secondary` full-width căn giữa bên dưới danh sách, label "Xem thêm", chuyển `loading` khi đang tải thêm, tự ẩn khi đã tải hết.

## Avatar

**Mục đích:** ảnh đại diện người dùng. **Dùng ở:** `Header` → `UserMenu`, profile.md, danh sách trong user-management.md.

**Cấu trúc:** hình tròn (`rounded-full`), size `sm` (24px, trong table) / `md` (32px, trong Header) / `lg` (80px, trong profile.md). Không có `avatar_url`: hiện chữ cái đầu tên trên nền `bg-blue-100 text-blue-700`.

## Tabs

**Mục đích:** chia nội dung 1 trang thành nhiều nhóm xem trong cùng ngữ cảnh. **Dùng ở:** event-manage.md (Tổng quan/Loại vé/Người mua vé (Phase 2)/Thống kê (Phase 2)), user-management.md (Phase 2) (Tất cả/Chờ duyệt organizer/Đã khoá).

**Cấu trúc:** hàng ngang, mỗi tab `px-4 py-2 text-sm font-medium border-b-2`; tab active: `border-blue-600 text-blue-600`; tab thường: `border-transparent text-gray-500 hover:text-gray-700`.

**Hành vi:** đổi tab không tải lại trang, cập nhật query param (`?tab=`) để giữ trạng thái khi refresh/chia sẻ link.

## StatTile

**Mục đích:** ô số liệu tổng quan (KPI) trong dashboard. **Dùng ở:** organizer/dashboard.md, admin/dashboard.md, tab Thống kê trong event-manage.md (tất cả: Phase 2).

**Cấu trúc:** khối `rounded-lg border border-gray-200 p-6`, nhãn phía trên (`Body small`, vd "Doanh thu"), số liệu lớn giữa (`text-3xl font-bold`), dòng nhỏ dưới cùng thể hiện thay đổi so với kỳ trước nếu có (`success`/`danger` kèm icon mũi tên lên/xuống).
