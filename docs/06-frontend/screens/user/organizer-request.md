# Đăng ký trở thành Organizer

**Route:** `/me/organizer-request` · **Vai trò:** user (role hiện tại = `user`) · **Render mode:** CSR · **Phase:** 2 (Must-have)

**Liên quan:** [docs/02-domains/identity/spec.md](../../02-domains/identity/spec.md) (mục "Đăng ký trở thành organizer") · [api-docs/openapi/identity-service.yaml](../../../api-docs/openapi/identity-service.yaml)

## Layout tổng quan

```
┌─────────────────────────────────────────┐
│ Header (variant public)                   │
├─────────────────────────────────────────┤
│ [Panel trạng thái / form đăng ký]         │
└─────────────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Header, `Button` | Component dùng chung |
| Panel trạng thái/form (nội dung thay đổi theo `users.status`) | Riêng màn hình này |

## Chi tiết section riêng

Nội dung panel (`max-w-lg mx-auto`) thay đổi theo trạng thái hiện tại của user:

| `users.status`/`role` | Nội dung hiển thị |
|---|---|
| `status=active`, `role=user` (chưa từng đăng ký) | Giới thiệu ngắn quyền lợi organizer + `Button` `primary` "Gửi yêu cầu" |
| `status=pending` | Icon đồng hồ `warning` + "Yêu cầu của bạn đang được xem xét. Chúng tôi sẽ thông báo khi có kết quả." — không có nút hành động |
| `role=organizer` (đã được duyệt) | Icon check `success` + "Bạn đã là Organizer!" + `Button` `primary` "Đến Kênh Organizer" → organizer/dashboard.md |

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Bấm "Gửi yêu cầu" | Gọi `POST /users/me/organizer-request` → thành công: cập nhật panel sang trạng thái `pending` ngay, `Toast` `success` "Đã gửi yêu cầu" |
| Gửi yêu cầu khi đã có yêu cầu `pending` (`409`) | `Toast` `danger`: "Bạn đã có yêu cầu đang chờ duyệt" — trường hợp này về lý thuyết không xảy ra vì UI đã ẩn nút khi `status=pending`, nhưng vẫn xử lý phòng trường hợp dữ liệu cũ trên client |

## Trạng thái đặc biệt

Không có empty/error state riêng ngoài 3 trạng thái nội dung đã liệt kê ở trên.

## Responsive

Không có khác biệt đáng kể.
