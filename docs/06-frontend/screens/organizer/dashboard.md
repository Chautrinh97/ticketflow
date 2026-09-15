# Dashboard Organizer

**Route:** `/organizer` · **Vai trò:** organizer · **Render mode:** CSR · **Phase:** 2 (Must-have)

**Liên quan:** [docs/02-domains/analytics/spec.md](../../02-domains/analytics/spec.md) · [api-docs/openapi/booking-service.yaml](../../../api-docs/openapi/booking-service.yaml) (`GET /organizer/stats`)

Dùng layout dashboard chung với [organizer/event-list.md](event-list.md), [organizer/event-form.md](event-form.md), [organizer/event-manage.md](event-manage.md): `Header` variant `dashboard` + `Sidebar` variant `organizer`.

## Layout tổng quan

```
┌───────┬─────────────────────────────────┐
│Sidebar│ Header (variant dashboard)        │
│       ├───────────────────────────────────┤
│ Tổng  │ [Hàng StatTile] Doanh thu | Vé đã  │
│ quan  │   bán | Tỉ lệ bán vé | Sự kiện     │
│ Sự    │   đang published                  │
│ kiện  ├───────────────────────────────────┤
│ của   │ [Danh sách] 5 sự kiện gần đây nhất │
│ tôi   │   (rút gọn từ event-list.md)       │
└───────┴─────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Header, Sidebar, `StatTile` | Component dùng chung — [components/layout.md](../../components/layout.md), [components/data-display.md](../../components/data-display.md#stattile) |
| Hàng StatTile (nội dung), Danh sách sự kiện gần đây | Riêng màn hình này |

## Chi tiết section riêng

- **Hàng StatTile**: 4 ô — "Doanh thu" (tổng `orders.total_amount` các đơn `paid`), "Vé đã bán", "Tỉ lệ bán vé" (`%`, `sold_count`/`quota` trung bình các sự kiện), "Sự kiện đang published" (đếm số lượng) — dữ liệu từ `GET /organizer/stats`.
- **Danh sách sự kiện gần đây**: 5 dòng gần nhất theo `created_at`, mỗi dòng: tên sự kiện, `Badge` trạng thái, số vé đã bán/tổng. Link "Xem tất cả" ở cuối → event-list.md.

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Click 1 dòng trong danh sách sự kiện gần đây | Điều hướng sang event-manage.md của sự kiện đó |
| Bấm "Xem tất cả" | Điều hướng sang event-list.md |

## Trạng thái đặc biệt

Organizer chưa tạo sự kiện nào: `StatTile` hiển thị 0 cho mọi ô, danh sách sự kiện gần đây thay bằng `EmptyState` — "Bạn chưa có sự kiện nào" + nút "Tạo sự kiện đầu tiên" (`primary`) → event-form.md (chế độ tạo mới).

## Responsive

Hàng `StatTile` chuyển 2 cột dưới `lg`, 1 cột dưới `sm`; `Sidebar` thu thành drawer dưới `lg` (xem [components/layout.md](../../components/layout.md#sidebar)).
