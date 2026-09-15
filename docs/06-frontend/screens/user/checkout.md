# Đặt vé / Checkout

**Route:** `/events/[slug]/checkout` · **Vai trò:** user (đã đăng nhập, mọi role đều đặt được vé — xem [docs/00-overview/roles-permissions.md](../../00-overview/roles-permissions.md)) · **Render mode:** CSR · **Phase:** 1 (MVP)

**Liên quan:** [docs/02-domains/booking/spec.md](../../02-domains/booking/spec.md) · [docs/02-domains/payment/spec.md](../../02-domains/payment/spec.md) · [api-docs/openapi/booking-service.yaml](../../../api-docs/openapi/booking-service.yaml) · [api-docs/openapi/payment-service.yaml](../../../api-docs/openapi/payment-service.yaml)

Đây là **1 trang duy nhất** với 2 bước hiển thị tuần tự (không phải 2 route riêng) — bước 2 chỉ hiện sau khi bước 1 hợp lệ.

## Layout tổng quan

```
┌─────────────────────────────────────────┐
│ Header (variant public, ẩn SearchBar)     │
├─────────────────────────────────────────┤
│ [Stepper] Bước 1: Chọn vé → Bước 2: Xác nhận │
├─────────────────────────────────────────┤
│ [Bước 1] Danh sách TicketTypeRow +        │
│          QuantityStepper mỗi loại vé      │
├─────────────────────────────────────────┤
│ [Bước 2] Tóm tắt đơn hàng + Tổng tiền     │
│          Nút "Tiến hành thanh toán"       │
└─────────────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Header | Component dùng chung — [components/layout.md](../../components/layout.md) |
| `TicketTypeRow` + `QuantityStepper` | Component dùng chung — [components/data-display.md](../../components/data-display.md#tickettyperow), [components/buttons-inputs.md](../../components/buttons-inputs.md#quantitystepper) |
| Stepper, Tóm tắt đơn hàng | Riêng màn hình này |

## Chi tiết section riêng

- **Stepper**: 2 mốc ngang ("1. Chọn vé", "2. Xác nhận"), mốc đã qua/đang ở → `text-blue-600`, mốc chưa tới → `text-gray-400`. Không cho click nhảy thẳng sang bước 2 nếu bước 1 chưa hợp lệ (chưa chọn vé nào).
- **Bước 1 — Chọn vé**: liệt kê toàn bộ `ticket_types` của sự kiện, mỗi dòng là `TicketTypeRow` (kèm `QuantityStepper`, `max` = `min(10, quota - sold_count)`). Cuối danh sách: dòng tổng tạm tính (`Body` "Tạm tính: <tổng tiền>"). Nút "Tiếp tục" (`primary`) — disable nếu tổng số vé đã chọn = 0.
- **Bước 2 — Xác nhận**: bảng tóm tắt (loại vé × số lượng × đơn giá × thành tiền), dòng "Tổng cộng" in đậm cuối bảng. Nút "Đổi lại" (`secondary`, quay về Bước 1) + nút "Tiến hành thanh toán" (`primary`).

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Đổi số lượng ở `QuantityStepper` | Cập nhật tạm tính ngay (tính toán phía client, không gọi API) |
| Bấm "Tiếp tục" (đủ điều kiện) | Chuyển sang hiển thị Bước 2 với dữ liệu đã chọn |
| Bấm "Tiến hành thanh toán" | Gọi `POST /bookings` (không optimistic — phải chờ response thật, xem [interaction-patterns.md](../../interaction-patterns.md#optimistic-update)) → thành công: gọi tiếp `POST /payments/:orderId/checkout` → redirect trình duyệt sang `checkout_url` của cổng thanh toán |
| `POST /bookings` trả `409` (không đủ tồn kho) | `Toast` `danger`: "Rất tiếc, số lượng vé bạn chọn không còn đủ. Vui lòng chọn lại." → quay về Bước 1, tự làm mới lại `quota`/`sold_count` hiện tại |
| `POST /bookings` trả `429` (vượt rate limit) | `Toast` `danger`: "Bạn thao tác quá nhanh, vui lòng thử lại sau vài giây" |
| Rời trang khi đang ở Bước 2 (đổi tab, back) | Không cần confirm — đơn hàng chỉ thật sự tạo khi bấm "Tiến hành thanh toán" thành công |

## Trạng thái đặc biệt

Nút "Tiến hành thanh toán" chuyển `loading` ngay khi bấm, disable toàn bộ Bước 2 trong lúc chờ (không cho sửa số lượng ngược lại Bước 1) cho tới khi có kết quả redirect hoặc lỗi.

## Responsive

Dưới `md`: tóm tắt tổng tiền hiện dạng thanh cố định đáy màn hình (`fixed bottom-0`) song song với nội dung danh sách vé, thay vì phải cuộn xuống cuối mới thấy tổng tiền.
