# Chi tiết vé / đơn hàng

**Route:** `/me/bookings/[id]` · **Vai trò:** user (chủ đơn hàng, hoặc `super_admin`) · **Render mode:** CSR · **Phase:** 1 (xem chi tiết + mã QR) → 2 (huỷ vé)

**Liên quan:** [docs/02-domains/booking/spec.md](../../02-domains/booking/spec.md) · [docs/04-security/authorization.md](../../04-security/authorization.md) · [api-docs/openapi/booking-service.yaml](../../../api-docs/openapi/booking-service.yaml)

## Layout tổng quan

```
┌─────────────────────────────────────────┐
│ Header (variant public)                   │
├─────────────────────────────────────────┤
│ [Thông tin sự kiện + Badge trạng thái]    │
├─────────────────────────────────────────┤
│ [Danh sách vé, mỗi vé 1 card có mã QR]    │
├─────────────────────────────────────────┤
│ [Thông tin thanh toán] [Nút "Huỷ đơn"]    │
└─────────────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Header, Badge trạng thái | Component dùng chung |
| Nút "Huỷ đơn" → mở `ConfirmDialog` | Component dùng chung — [components/feedback.md](../../components/feedback.md#confirmdialog) |
| Thông tin sự kiện, danh sách vé QR, thông tin thanh toán | Riêng màn hình này |

## Chi tiết section riêng

- **Thông tin sự kiện**: ảnh thumbnail + tên sự kiện (link sang event-detail.md) + thời gian/địa điểm + `Badge` trạng thái đơn hàng (`orders.status`).
- **Danh sách vé**: mỗi `tickets` record 1 card riêng (`rounded-lg border p-6 text-center`) — chỉ hiển thị nếu `orders.status = paid`: mã QR lớn ở giữa (sinh từ `tickets.ticket_code`), tên loại vé bên dưới, `ticket_code` dạng text nhỏ (`Caption`, để soát vé thủ công nếu cần), `Badge` trạng thái vé (`valid`/`used`/`cancelled`). Nếu `orders.status != paid`: thay bằng thông báo "Vé sẽ hiển thị sau khi thanh toán thành công".
- **Thông tin thanh toán**: bảng nhỏ liệt kê từng `order_items` (loại vé × số lượng × đơn giá) + tổng tiền + thời gian tạo đơn (`created_at`).

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Bấm "Huỷ đơn" (chỉ hiện khi `orders.status = paid` và còn trong hạn cho phép huỷ) | Mở `ConfirmDialog` nội dung "Vé sẽ bị huỷ và không thể khôi phục. Tiếp tục?" |
| Xác nhận trong `ConfirmDialog` | Gọi `POST /bookings/:id/cancel` → thành công: cập nhật `Badge` trạng thái ngay tại chỗ + `Toast` `success` "Đã huỷ đơn hàng" |
| Huỷ thất bại (`409`, quá hạn cho phép huỷ) | `Toast` `danger`: "Đã quá thời hạn cho phép huỷ vé" |
| Bấm vào ảnh/tên sự kiện | Điều hướng sang event-detail.md |

## Trạng thái đặc biệt

`id` không tồn tại hoặc không thuộc user hiện tại: màn hình lỗi 404 — xem [system/error-states.md](../system/error-states.md).

## Responsive

Không có khác biệt đáng kể — layout vốn 1 cột.
