# Kết quả thanh toán

**Route:** `/orders/[orderId]/confirmation` · **Vai trò:** user (chủ đơn hàng) · **Render mode:** CSR · **Phase:** 1 (MVP)

**Liên quan:** [docs/02-domains/booking/spec.md](../../02-domains/booking/spec.md) · [docs/02-domains/payment/spec.md](../../02-domains/payment/spec.md) · [api-docs/openapi/booking-service.yaml](../../../api-docs/openapi/booking-service.yaml)

Trang người dùng được cổng thanh toán redirect về sau khi hoàn tất (hoặc huỷ) thanh toán. Kết quả thật sự do webhook xác nhận ở backend quyết định (xem [docs/02-domains/payment/spec.md](../../02-domains/payment/spec.md)) — trang này chỉ **hiển thị** trạng thái hiện tại của đơn hàng, không tự quyết định thành công/thất bại.

## Layout tổng quan

```
┌─────────────────────────────────────────┐
│ Header (variant public)                   │
├─────────────────────────────────────────┤
│         [Icon trạng thái lớn]             │
│         [Tiêu đề kết quả]                 │
│         [Mô tả ngắn]                      │
│         [Tóm tắt đơn hàng]                │
│         [Nút hành động]                   │
└─────────────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Header | Component dùng chung |
| Toàn bộ nội dung còn lại | Riêng màn hình này |

## Chi tiết section riêng

Nội dung khối trung tâm (`max-w-lg mx-auto text-center py-16`) thay đổi theo `orders.status`:

| `orders.status` | Icon | Tiêu đề | Mô tả | Nút hành động |
|---|---|---|---|---|
| `paid` | check tròn `success` | "Đặt vé thành công!" | "Vé của bạn đã sẵn sàng, kiểm tra mã QR trong mục Vé của tôi." | "Xem vé của tôi" (`primary`) → booking-detail.md |
| `pending` | đồng hồ `warning` | "Đang xác nhận thanh toán" | "Hệ thống đang xác nhận giao dịch, vui lòng đợi trong giây lát." | tự động polling (xem Hành vi tương tác) |
| `expired` / `cancelled` | dấu X tròn `danger` | "Thanh toán không thành công" | "Đơn hàng đã bị huỷ hoặc hết hạn thanh toán. Vé đã được hoàn lại vào kho." | "Đặt lại vé" (`primary`) → event-detail.md của sự kiện tương ứng |

Bên dưới luôn có **Tóm tắt đơn hàng** thu gọn: tên sự kiện, loại vé × số lượng, tổng tiền (dùng lại cấu trúc bảng tóm tắt ở checkout.md Bước 2, chỉ để xem).

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Vào trang, `orders.status = pending` | Gọi `GET /bookings/:id` lặp lại mỗi 3 giây (polling) cho tới khi `status` chuyển `paid`/`expired`/`cancelled`, tối đa 2 phút — quá thời gian này vẫn `pending`: đổi mô tả thành "Việc xác nhận đang mất nhiều thời gian hơn dự kiến, bạn có thể kiểm tra lại trong mục Vé của tôi" + nút "Xem Vé của tôi" |
| `status` chuyển `paid` trong lúc đang polling | Dừng polling, cập nhật UI sang trạng thái thành công ngay (không cần reload) |

## Trạng thái đặc biệt

`orderId` không tồn tại hoặc không thuộc user hiện tại: hiển thị màn hình lỗi 404 — xem [system/error-states.md](../system/error-states.md).

## Responsive

Không có khác biệt đáng kể — layout vốn đã là 1 cột căn giữa.
