# Vé của tôi

**Route:** `/me/bookings` · **Vai trò:** user (đã đăng nhập) · **Render mode:** CSR · **Phase:** 1 (MVP)

**Liên quan:** [docs/02-domains/booking/spec.md](../../02-domains/booking/spec.md) · [api-docs/openapi/booking-service.yaml](../../../api-docs/openapi/booking-service.yaml)

## Layout tổng quan

```
┌─────────────────────────────────────────┐
│ Header (variant public)                   │
├─────────────────────────────────────────┤
│ [Tabs] Tất cả | Chờ thanh toán | Đã xong │
│        | Đã huỷ                          │
├─────────────────────────────────────────┤
│ [Danh sách đơn hàng dạng card, dọc]       │
│   [Nút "Xem thêm"]                        │
└─────────────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Header, Tabs | Component dùng chung — [components/layout.md](../../components/layout.md), [`Tabs`](../../components/data-display.md#tabs) |
| `Badge` trạng thái trong mỗi card | Component dùng chung — [components/data-display.md](../../components/data-display.md#badge--statustag) |
| Danh sách đơn hàng (bố cục card) | Riêng màn hình này |

## Chi tiết section riêng

Mỗi item là 1 card ngang (`rounded-lg border border-gray-200 p-4 flex`): ảnh nhỏ sự kiện (banner thumbnail 80×80) → tên sự kiện + thời gian diễn ra (`Body small`) → số lượng vé đã mua → `Badge` trạng thái đơn hàng → giá trị đơn hàng (`font-semibold`, căn phải). Cả card là link tới booking-detail.md.

`Tabs` map trực tiếp với `orders.status`: "Tất cả" (không filter) · "Chờ thanh toán" (`pending`) · "Đã xong" (`paid`) · "Đã huỷ" (`cancelled`, `expired` gộp chung).

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Đổi tab | Cập nhật query `?status=`, gọi lại `GET /users/me/bookings` với filter tương ứng |
| Click 1 card | Điều hướng sang booking-detail.md |
| Bấm "Xem thêm" | Tải thêm trang kế tiếp, nối cuối danh sách (quy ước "Xem thêm") |

## Trạng thái đặc biệt

Tab "Tất cả" rỗng (chưa từng đặt vé): `EmptyState` — "Bạn chưa có đơn vé nào" + nút "Khám phá sự kiện" (`primary`) → home.md. Các tab khác rỗng: text mô tả tương ứng, không có CTA (vd tab "Đã huỷ" rỗng → "Bạn chưa huỷ đơn nào").

## Responsive

Card chuyển bố cục dọc dưới `sm` (ảnh trên, thông tin dưới) thay vì ngang.
