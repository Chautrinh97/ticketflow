# Vai trò người dùng & phân quyền

## Vai trò

| Vai trò | Mô tả | Phạm vi quyền |
|---|---|---|
| `super_admin` | Quản trị viên nền tảng | Toàn quyền: quản lý user, duyệt/khoá organizer, xem thống kê toàn hệ thống, xem audit log |
| `organizer` | Đơn vị tổ chức sự kiện | Chỉ quản lý được sự kiện **do chính mình tạo** (ownership-based, không chỉ role-based) |
| `user` | Người mua vé | Xem, tìm kiếm, đặt vé, quản lý vé/đơn hàng của chính mình |

## Ma trận quyền theo hành động

| Hành động | super_admin | organizer | user |
|---|:---:|:---:|:---:|
| Quản lý tài khoản user khác | ✅ | ❌ | ❌ |
| Duyệt / khoá tài khoản organizer | ✅ | ❌ | ❌ |
| Tạo / sửa / xoá sự kiện | ✅ (mọi sự kiện) | ✅ (chỉ sự kiện sở hữu) | ❌ |
| Xem báo cáo doanh thu | ✅ (toàn hệ thống) | ✅ (chỉ sự kiện của mình) | ❌ |
| Đặt vé | ✅ | ✅ | ✅ |
| Huỷ vé đã mua | ✅ | ❌ | ✅ (vé của chính mình) |
| Xem audit log | ✅ | ❌ | ❌ |

## Nguyên tắc thiết kế authorization

Middleware phân quyền **không được chỉ kiểm tra `role == organizer`**. Với mọi hành động thao tác lên một resource cụ thể (sự kiện, đơn hàng, vé), còn phải kiểm tra **ownership**: `resource.owner_id == current_user.id` (vd: `event.organizer_id == current_user.id` khi organizer sửa/xoá sự kiện, `order.user_id == current_user.id` khi user huỷ vé). Đây là phần dễ bị làm sơ sài — chỉ kiểm role mà bỏ qua ownership sẽ cho phép một organizer sửa/xoá sự kiện của organizer khác.

Chi tiết cơ chế middleware xem tại [../04-security/authorization.md](../04-security/authorization.md).

## Vòng đời trở thành organizer

Tài khoản `user` đăng ký trở thành `organizer` phải qua bước **duyệt bởi `super_admin`** — không tự động nâng quyền. Chi tiết luồng xem tại [../02-domains/identity/spec.md](../02-domains/identity/spec.md).
