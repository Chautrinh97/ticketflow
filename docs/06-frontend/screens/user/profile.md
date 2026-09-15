# Thông tin cá nhân

**Route:** `/me/profile` · **Vai trò:** user (mọi role đã đăng nhập) · **Render mode:** CSR · **Phase:** 1 (chỉnh sửa họ tên) → 2 (avatar upload)

**Liên quan:** [docs/02-domains/identity/spec.md](../../02-domains/identity/spec.md) · [api-docs/openapi/identity-service.yaml](../../../api-docs/openapi/identity-service.yaml)

## Layout tổng quan

```
┌─────────────────────────────────────────┐
│ Header (variant public)                   │
├─────────────────────────────────────────┤
│ [Panel thông tin cá nhân]                 │
│  - Avatar + FileUpload (Phase 2)          │
│  - Input họ tên (email chỉ đọc)           │
│  - Nút "Lưu thay đổi"                     │
└─────────────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Header, `Avatar`, `FileUpload`, `Input`, `Button` | Component dùng chung — [components/layout.md](../../components/layout.md), [components/data-display.md](../../components/data-display.md#avatar), [components/buttons-inputs.md](../../components/buttons-inputs.md) |
| Panel thông tin cá nhân (bố cục) | Riêng màn hình này |

## Chi tiết section riêng

Panel `max-w-lg rounded-lg border border-gray-200 p-6`:

- `FileUpload` variant tròn (avatar), size `lg` (Phase 2) — đè icon máy ảnh nhỏ góc dưới phải khi hover để gợi ý đổi ảnh.
- `Input` họ tên (`full_name`, chỉnh sửa được).
- `Input` email (`readonly` — email gắn với tài khoản Firebase, không cho sửa trực tiếp ở đây).
- `Button` variant `primary` "Lưu thay đổi", disable nếu chưa có thay đổi nào so với dữ liệu ban đầu.

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Chọn ảnh mới qua `FileUpload` (Phase 2) | Upload theo luồng presigned URL (xem [components/buttons-inputs.md](../../components/buttons-inputs.md#fileupload--imageuploader)); sau khi có `public_url`, kích hoạt nút "Lưu thay đổi" |
| Sửa họ tên | Kích hoạt nút "Lưu thay đổi" |
| Bấm "Lưu thay đổi" | Validate theo quy tắc chung → gọi `PATCH /users/me` → thành công: `Toast` `success` "Đã cập nhật thông tin"; thất bại: `Toast` `danger` |

## Trạng thái đặc biệt

Loading dữ liệu ban đầu: `Skeleton` cho avatar (Phase 2) + 2 dòng input.

## Responsive

Panel full-width trừ padding dưới `md`.
