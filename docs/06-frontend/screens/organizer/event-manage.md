# Quản lý sự kiện (chi tiết)

**Route:** `/organizer/events/[id]` · **Vai trò:** organizer (chỉ chủ sở hữu — ownership check) · **Render mode:** CSR · **Phase:** 1 (xem + thêm loại vé + xuất bản) → 2 (sửa/huỷ sự kiện, sửa nhanh loại vé, tab Người mua vé/Thống kê)

**Liên quan:** [docs/02-domains/event-catalog/spec.md](../../02-domains/event-catalog/spec.md) · [docs/02-domains/booking/spec.md](../../02-domains/booking/spec.md) · [docs/02-domains/analytics/spec.md](../../02-domains/analytics/spec.md) · [api-docs/openapi/event-service.yaml](../../../api-docs/openapi/event-service.yaml) · [api-docs/openapi/booking-service.yaml](../../../api-docs/openapi/booking-service.yaml) (`GET .../buyers`, `GET .../stats`)

Dùng layout dashboard chung — xem [organizer/dashboard.md](dashboard.md#layout-tổng-quan).

## Layout tổng quan

```
┌───────┬─────────────────────────────────┐
│Sidebar│ Header (variant dashboard)        │
│       ├───────────────────────────────────┤
│       │ [Tiêu đề sự kiện + Badge + Button  │
│       │   "Sửa"/"Xuất bản"/"Huỷ sự kiện"]  │
│       ├───────────────────────────────────┤
│       │ [Tabs] Tổng quan|Loại vé|Người mua │
│       │   vé|Thống kê                      │
│       ├───────────────────────────────────┤
│       │ [Nội dung theo tab đang chọn]      │
└───────┴─────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Header, Sidebar, `Tabs`, `Badge`, `Button`, `ConfirmDialog`, `DataTable`, `StatTile` | Component dùng chung |
| Tiêu đề + nút hành động, nội dung từng tab | Riêng màn hình này |

## Chi tiết section riêng

- **Tiêu đề**: tên sự kiện (`H1`) + `Badge` trạng thái cạnh bên; bên phải: `Button` `secondary` "Sửa" (Phase 2) → event-form.md (chế độ chỉnh sửa), `Button` `primary` "Xuất bản" (chỉ hiện khi `status=draft`), `Button` `danger` ghost "Huỷ sự kiện" (Phase 2, chỉ hiện khi `status != cancelled`).
- **Tab Tổng quan**: hiển thị lại thông tin cơ bản đã nhập ở event-form.md (chỉ xem — không sửa trực tiếp ở đây).
- **Tab Loại vé**: bảng các `ticket_types` (tên, giá, `sold_count`/`quota`), `IconButton` sửa mở `Modal` sửa nhanh 1 loại vé (tên/giá/quota) (Phase 2), nút "+ Thêm loại vé" mở cùng `Modal` ở chế độ tạo mới.
- **Tab Người mua vé** (Phase 2): `DataTable` (tên người mua, email, loại vé, số lượng, trạng thái đơn hàng, thời gian đặt) từ `GET /organizer/events/:id/buyers`, phân trang theo số trang.
- **Tab Thống kê** (Phase 2): hàng `StatTile` (Doanh thu sự kiện này, Tổng vé đã bán, Tỉ lệ bán vé) từ `GET /organizer/events/:id/stats`, bên dưới là bảng chi tiết theo từng loại vé (số lượng bán/doanh thu riêng từng loại).

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Bấm "Xuất bản" | Gọi `POST /organizer/events/:id/publish` → thành công: cập nhật `Badge` ngay, `Toast` `success` |
| Bấm "Huỷ sự kiện" (Phase 2) | Mở `ConfirmDialog` (nội dung theo [interaction-patterns.md](../../interaction-patterns.md#confirm-dialog-modal-xác-nhận-hành-động-nguy-hiểm)) → xác nhận: gọi `DELETE /organizer/events/:id` → thành công: cập nhật `Badge` sang `cancelled`, `Toast` `success` |
| Huỷ sự kiện thất bại (`409`, đã có đơn `paid`) (Phase 2) | `Toast` `danger`: "Sự kiện đã có người mua vé, không thể xoá hẳn — sự kiện sẽ được chuyển sang trạng thái Đã huỷ thay vì xoá dữ liệu" (đối chiếu domain spec: huỷ = đổi `status='cancelled'`, không xoá cứng) |
| Đổi tab | Cập nhật query `?tab=`, tải dữ liệu tab tương ứng nếu chưa có sẵn |
| Sửa nhanh 1 loại vé trong `Modal` (Phase 2) | Gọi API cập nhật ticket type tương ứng → thành công: đóng `Modal`, cập nhật bảng, `Toast` `success` |

## Trạng thái đặc biệt

Tab Người mua vé chưa có ai mua: `EmptyState` — "Chưa có ai mua vé cho sự kiện này".

## Responsive

`Tabs` cho cuộn ngang dưới `sm` nếu không đủ chỗ hiển thị hết 4 tab cùng lúc.
