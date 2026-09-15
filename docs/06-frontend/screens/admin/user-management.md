# Quản lý người dùng

**Route:** `/admin/users` · **Vai trò:** `super_admin` · **Render mode:** CSR · **Phase:** 2 (Must-have)

**Liên quan:** [docs/00-overview/roles-permissions.md](../../00-overview/roles-permissions.md) · [docs/02-domains/identity/spec.md](../../02-domains/identity/spec.md) · [api-docs/openapi/identity-service.yaml](../../../api-docs/openapi/identity-service.yaml)

Gồm cả duyệt yêu cầu organizer và khoá/mở khoá tài khoản trong cùng 1 màn hình (qua `Tabs`), không tách route riêng. Dùng layout dashboard chung — xem [admin/dashboard.md](dashboard.md#layout-tổng-quan).

## Layout tổng quan

```
┌───────┬─────────────────────────────────┐
│Sidebar│ Header (variant dashboard)        │
│       ├───────────────────────────────────┤
│       │ [Tabs] Tất cả | Chờ duyệt         │
│       │   organizer | Đã khoá             │
│       ├───────────────────────────────────┤
│       │ [Thanh trên] SearchBar theo email/ │
│       │   tên + Select role               │
│       ├───────────────────────────────────┤
│       │ [DataTable] User|Role|Trạng thái|  │
│       │   Ngày tạo|Hành động               │
│       │ [Pagination]                       │
└───────┴─────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Header, Sidebar, `Tabs`, `DataTable`, `Pagination`, `Select`, `Badge`, `ConfirmDialog` | Component dùng chung |
| Thanh trên, nội dung cột "Hành động" theo từng tab | Riêng màn hình này |

## Chi tiết section riêng

- **`Tabs`**: "Tất cả" (không filter) · "Chờ duyệt organizer" (`status=pending`) · "Đã khoá" (`status=banned`).
- **Thanh trên**: `SearchBar` biến thể nhỏ (tìm theo `email`/`full_name`) + `Select` filter theo `role`.
- **`DataTable`**: cột "User" (`Avatar` size `sm` + tên + email), "Role" (`Badge` neutral, không phải trạng thái nên không dùng màu success/danger), "Trạng thái" (`Badge` theo bảng map ở [components/data-display.md](../../components/data-display.md#badge--statustag)), "Ngày tạo". Cột "Hành động" thay đổi theo tab:
  - Tab "Chờ duyệt organizer": 2 `Button` nhỏ "Duyệt" (`primary`)/"Từ chối" (`danger` ghost).
  - Tab "Tất cả"/"Đã khoá": 1 `Button` nhỏ "Khoá tài khoản" (nếu đang `active`) hoặc "Mở khoá" (nếu đang `banned`) — không hiện với chính tài khoản `super_admin` đang đăng nhập.

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Bấm "Duyệt" | Gọi `PATCH /admin/users/:id/role` (`role=organizer`, `status=active`) → cập nhật dòng ngay, `Toast` `success` |
| Bấm "Từ chối" | `ConfirmDialog` → xác nhận: `PATCH /admin/users/:id/role` (`status=active`, giữ `role=user`) → cập nhật dòng, `Toast` `success` |
| Bấm "Khoá tài khoản" | `ConfirmDialog` (nội dung theo [interaction-patterns.md](../../interaction-patterns.md#confirm-dialog-modal-xác-nhận-hành-động-nguy-hiểm)) → xác nhận: `PATCH /admin/users/:id/role` (`status=banned`) → cập nhật `Badge`, `Toast` `success` |
| Bấm "Mở khoá" | `PATCH /admin/users/:id/role` (`status=active`) → cập nhật `Badge`, `Toast` `success` (không cần `ConfirmDialog` vì mở khoá không phải hành động phá huỷ) |
| Gõ vào `SearchBar`/đổi `Select` role | Debounce 300ms rồi gọi lại `GET /admin/users` với filter tương ứng |

## Trạng thái đặc biệt

Tab "Chờ duyệt organizer" rỗng: `EmptyState` — "Không có yêu cầu nào đang chờ duyệt".

## Responsive

`DataTable` cho phép cuộn ngang dưới `md`.
