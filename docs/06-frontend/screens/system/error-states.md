# Trạng thái lỗi hệ thống

**Route:** áp dụng cho mọi route khi gặp lỗi tương ứng (không phải 1 route cố định) · **Vai trò:** mọi vai trò · **Render mode:** 404 dùng SSR (để trả đúng HTTP status code cho SEO/crawler); lỗi runtime dùng CSR (React error boundary) · **Phase:** 1 (MVP)

Đây là các trạng thái **toàn trang**, khác với "Trạng thái đặc biệt" mô tả trong từng screen spec (vốn chỉ thay 1 phần nội dung, không phải toàn trang) và khác với `EmptyState`/`Toast` (mô tả ở [components/feedback.md](../../components/feedback.md)).

## Layout tổng quan

```
┌─────────────────────────────────────────┐
│ Header (variant public, nếu còn xác định  │
│  được — nếu lỗi ngay từ layout thì ẩn)    │
├─────────────────────────────────────────┤
│            [Icon/hình minh hoạ lớn]       │
│            [Tiêu đề lỗi]                  │
│            [Mô tả ngắn]                   │
│            [Nút hành động]                │
└─────────────────────────────────────────┘
```

Toàn bộ nội dung khối trung tâm là riêng của từng loại lỗi — không có component dùng chung ngoài `Header`/`Button`.

## Chi tiết section riêng

| Loại lỗi | Khi nào | Tiêu đề | Mô tả | Nút hành động |
|---|---|---|---|---|
| 404 Not Found | Route không khớp, hoặc resource không tồn tại/không thuộc quyền xem (event/booking/user không tìm thấy) | "Không tìm thấy trang" | "Trang bạn tìm không tồn tại hoặc đã bị gỡ bỏ." | "Về trang chủ" (`primary`) → home.md |
| 403 Forbidden | Người dùng cố truy cập route yêu cầu role không đúng (vd `user` vào `/admin`) | "Bạn không có quyền truy cập" | "Trang này chỉ dành cho quản trị viên/organizer." | "Về trang chủ" (`primary`) |
| Lỗi runtime (error boundary) | Exception không lường trước ở phía client | "Đã có lỗi xảy ra" | "Rất tiếc, đã có lỗi ngoài dự kiến. Vui lòng thử lại." | "Tải lại trang" (`primary`) |
| Mất kết nối mạng | `fetch` thất bại do offline | "Mất kết nối mạng" | "Kiểm tra kết nối Internet và thử lại." | "Thử lại" (`primary`) |

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Bấm nút hành động (tuỳ loại lỗi) | Điều hướng về trang chủ hoặc thử tải lại request/trang tương ứng |

## Trạng thái đặc biệt

Không áp dụng — bản thân file này mô tả các trạng thái đặc biệt toàn trang.

## Responsive

Layout vốn 1 cột căn giữa — không có khác biệt đáng kể.
