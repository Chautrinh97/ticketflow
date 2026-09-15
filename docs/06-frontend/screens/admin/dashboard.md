# Dashboard Admin

**Route:** `/admin` · **Vai trò:** `super_admin` · **Render mode:** CSR · **Phase:** 2 (Must-have)

**Liên quan:** [docs/02-domains/analytics/spec.md](../../02-domains/analytics/spec.md) · [api-docs/openapi/booking-service.yaml](../../../api-docs/openapi/booking-service.yaml) (`GET /admin/stats`)

Dùng layout dashboard chung với [admin/user-management.md](user-management.md), [admin/audit-log.md](audit-log.md): `Header` variant `dashboard` + `Sidebar` variant `admin`.

## Layout tổng quan

```
┌───────┬─────────────────────────────────┐
│Sidebar│ Header (variant dashboard)        │
│       ├───────────────────────────────────┤
│ Tổng  │ [Hàng StatTile] Tổng doanh thu |   │
│ quan  │   Tổng đơn hàng | Tổng user |      │
│ Người │   Tổng organizer                   │
│ dùng  ├───────────────────────────────────┤
│ Audit │ [Danh sách] Yêu cầu organizer đang │
│ log   │   chờ duyệt (rút gọn)               │
└───────┴─────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Header, Sidebar, `StatTile` | Component dùng chung |
| Hàng StatTile (nội dung), Danh sách yêu cầu chờ duyệt | Riêng màn hình này |

## Chi tiết section riêng

- **Hàng StatTile**: "Tổng doanh thu" (toàn hệ thống), "Tổng đơn hàng", "Tổng user", "Tổng organizer" — từ `GET /admin/stats`.
- **Danh sách yêu cầu chờ duyệt**: tối đa 5 dòng gần nhất có `users.status=pending`, mỗi dòng: tên + email + thời gian gửi yêu cầu + 2 `Button` nhỏ "Duyệt"/"Từ chối" ngay tại dòng. Link "Xem tất cả" → user-management.md (tab "Chờ duyệt organizer").

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Bấm "Duyệt" tại 1 dòng | Gọi `PATCH /admin/users/:id/role` (`role=organizer`, `status=active`) → thành công: xoá dòng khỏi danh sách, `Toast` `success` "Đã duyệt <tên>" |
| Bấm "Từ chối" tại 1 dòng | Mở `ConfirmDialog` (xem [interaction-patterns.md](../../interaction-patterns.md#confirm-dialog-modal-xác-nhận-hành-động-nguy-hiểm)) → xác nhận: gọi `PATCH /admin/users/:id/role` (`status=active`, giữ `role=user`) → xoá dòng khỏi danh sách, `Toast` `success` |
| Bấm "Xem tất cả" | Điều hướng sang user-management.md |

## Trạng thái đặc biệt

Không có yêu cầu nào đang chờ: ẩn hẳn section này (không hiện `EmptyState`, vì đây là trạng thái bình thường/tích cực, không cần nhấn mạnh).

## Responsive

Hàng `StatTile` chuyển 2 cột dưới `lg`, 1 cột dưới `sm`.
