# Sự kiện của tôi

**Route:** `/organizer/events` · **Vai trò:** organizer (chỉ thấy sự kiện của chính mình — ownership check, xem [docs/04-security/authorization.md](../../04-security/authorization.md)) · **Render mode:** CSR · **Phase:** 2 (Must-have)

**Liên quan:** [docs/02-domains/event-catalog/spec.md](../../02-domains/event-catalog/spec.md) · [api-docs/openapi/event-service.yaml](../../../api-docs/openapi/event-service.yaml) (`GET /organizer/events`)

Dùng layout dashboard chung — xem [organizer/dashboard.md](dashboard.md#layout-tổng-quan).

## Layout tổng quan

```
┌───────┬─────────────────────────────────┐
│Sidebar│ Header (variant dashboard)        │
│       ├───────────────────────────────────┤
│       │ [Thanh trên] Select trạng thái +   │
│       │   Button "Tạo sự kiện"             │
│       ├───────────────────────────────────┤
│       │ [DataTable] Sự kiện | Trạng thái | │
│       │   Thời gian | Vé đã bán | Hành động│
│       │ [Pagination]                       │
└───────┴─────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Header, Sidebar, `DataTable`, `Pagination`, `Select`, `Button` | Component dùng chung |
| Thanh trên (bố cục) | Riêng màn hình này |

## Chi tiết section riêng

- **Thanh trên**: `Select` filter theo `status` (Tất cả/`draft`/`published`/`cancelled`) bên trái, `Button` variant `primary` "Tạo sự kiện" bên phải → event-form.md (chế độ tạo mới).
- **`DataTable`**: cột "Sự kiện" (thumbnail nhỏ + tên), "Trạng thái" (`Badge`), "Thời gian" (`start_time`), "Vé đã bán" (`sold_count`/`quota` tổng các ticket_type), cột hành động cuối (`IconButton` xem chi tiết → event-manage.md, `IconButton` sửa → event-form.md chế độ chỉnh sửa).

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Đổi `Select` trạng thái | Gọi lại `GET /organizer/events?status=` |
| Bấm "Tạo sự kiện" | Điều hướng sang event-form.md (chế độ tạo mới) |
| Click 1 dòng (ngoài cột hành động) | Điều hướng sang event-manage.md |
| Đổi trang `Pagination` | Gọi lại `GET /organizer/events` với `page` tương ứng |

## Trạng thái đặc biệt

Chưa có sự kiện nào (filter "Tất cả"): `EmptyState` — "Bạn chưa có sự kiện nào" + nút "Tạo sự kiện đầu tiên" → event-form.md.

## Responsive

`DataTable` cho phép cuộn ngang (`overflow-x-auto`) dưới `md` thay vì ẩn bớt cột.
