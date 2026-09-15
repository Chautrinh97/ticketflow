# Audit log

**Route:** `/admin/audit-logs` · **Vai trò:** `super_admin` · **Render mode:** CSR · **Phase:** 2 (Must-have)

**Liên quan:** [docs/04-security/authorization.md](../../04-security/authorization.md) (mục "Audit log") · [docs/03-data/postgres-schema.md](../../03-data/postgres-schema.md) (bảng `audit_logs`) · [api-docs/openapi/identity-service.yaml](../../../api-docs/openapi/identity-service.yaml) (`GET /admin/audit-logs`)

Dùng layout dashboard chung — xem [admin/dashboard.md](dashboard.md#layout-tổng-quan). Đây là màn hình **chỉ đọc** — không có hành động ghi dữ liệu nào trên chính màn hình này.

## Layout tổng quan

```
┌───────┬─────────────────────────────────┐
│Sidebar│ Header (variant dashboard)        │
│       ├───────────────────────────────────┤
│       │ [Thanh trên] DatePicker khoảng     │
│       │   ngày + Select loại hành động     │
│       ├───────────────────────────────────┤
│       │ [DataTable] Thời gian|Người thực   │
│       │   hiện|Hành động|Đối tượng|Chi tiết│
│       │ [Pagination]                       │
└───────┴─────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Header, Sidebar, `DataTable`, `Pagination`, `DatePicker`, `Select` | Component dùng chung |
| Thanh trên, cột "Chi tiết" | Riêng màn hình này |

## Chi tiết section riêng

- **Thanh trên**: `DatePicker` khoảng ngày lọc `created_at`, `Select` lọc theo `action` (vd "Duyệt organizer", "Khoá tài khoản", "Huỷ sự kiện"...).
- **`DataTable`**: cột "Thời gian" (`created_at`, định dạng đầy đủ), "Người thực hiện" (`actor_id` → hiển thị tên nếu tra được, fallback hiển thị UUID rút gọn), "Hành động" (`action`, dạng text người đọc được, vd `user.role_changed` → "Đổi role người dùng"), "Đối tượng" (`resource_type` + `resource_id` rút gọn), "Chi tiết" — `IconButton` mở `Modal` xem toàn bộ `metadata` (dạng JSON được format đẹp, chỉ đọc).

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Đổi `DatePicker`/`Select` | Gọi lại `GET /admin/audit-logs` với filter tương ứng |
| Bấm `IconButton` "Chi tiết" ở 1 dòng | Mở `Modal` hiển thị `metadata` dạng JSON format đẹp (chỉ đọc, không có hành động ghi) |

## Trạng thái đặc biệt

Không có log nào khớp filter: `EmptyState` — "Không có log nào trong khoảng thời gian đã chọn".

## Responsive

`DataTable` cho phép cuộn ngang dưới `md`; cột "Chi tiết" luôn hiển thị (không bị ẩn) vì là cách duy nhất xem đầy đủ thông tin.
