# Thông báo

**Route:** `/me/notifications` · **Vai trò:** user (đã đăng nhập) · **Render mode:** CSR · **Phase:** 2 (Must-have)

**Liên quan:** [docs/02-domains/notification/spec.md](../../02-domains/notification/spec.md) · [api-docs/openapi/notification-service.yaml](../../../api-docs/openapi/notification-service.yaml)

Trang danh sách đầy đủ — bản rút gọn (5 mục gần nhất) hiển thị trong dropdown `NotificationBell` ở `Header`, xem [components/layout.md](../../components/layout.md#header).

## Layout tổng quan

```
┌─────────────────────────────────────────┐
│ Header (variant public)                   │
├─────────────────────────────────────────┤
│ [Thanh trên] "Đánh dấu tất cả đã đọc"     │
├─────────────────────────────────────────┤
│ [Danh sách thông báo, dọc]                │
│   [Nút "Xem thêm"]                        │
└─────────────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Header | Component dùng chung |
| Danh sách thông báo (bố cục item) | Riêng màn hình này |

## Chi tiết section riêng

Mỗi item (`flex items-start gap-3 py-4 border-b border-gray-100`): chấm tròn nhỏ bên trái (hiện nếu `is_read=false`, màu `bg-blue-600`) → icon theo `type` (vd `ticket.issued` → icon vé, `payment.failed` → icon cảnh báo) → nội dung: `title` (`font-medium`, đậm hơn nếu chưa đọc) + `body` (`Body small`) + thời gian tương đối (`Caption`, vd "2 giờ trước"). Item chưa đọc có nền `bg-blue-50/40` nhẹ.

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Click 1 thông báo | Gọi `PATCH /users/me/notifications/:id/read` (optimistic — cập nhật `is_read=true` trên UI ngay, xem [interaction-patterns.md](../../interaction-patterns.md#optimistic-update)), sau đó điều hướng tới màn hình liên quan nếu thông báo có ngữ cảnh cụ thể (vd `ticket.issued` → booking-detail.md tương ứng) |
| Bấm "Đánh dấu tất cả đã đọc" | Optimistic cập nhật toàn bộ item hiện có sang đã đọc, gọi API tương ứng cho từng item chưa đọc |
| Bấm "Xem thêm" | Tải thêm trang kế tiếp, nối cuối danh sách |

## Trạng thái đặc biệt

Chưa có thông báo nào: `EmptyState` — "Bạn chưa có thông báo nào".

## Responsive

Không có khác biệt đáng kể.
