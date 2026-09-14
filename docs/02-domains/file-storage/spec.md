# Domain: File Storage

**Service sở hữu:** File Service · **Database:** không có (stateless) · **Phase:** Phase 2 (Must-have).

## Phạm vi & trách nhiệm

Sinh presigned URL để client upload file (banner sự kiện, avatar user) trực tiếp lên object storage (DigitalOcean Spaces / AWS S3 hoặc tương đương), không cho file đi qua backend — giảm tải băng thông xử lý file cho hệ thống và tận dụng CDN có sẵn của storage provider.

## Luồng nghiệp vụ

```
1. Client gọi POST /files/presign { file_name, content_type }
2. File Service validate loại file/kích thước cho phép → sinh presigned URL (PUT) từ object storage
3. Client upload trực tiếp file lên URL đó (không qua backend)
4. Client gửi URL công khai về để lưu vào events.banner_url hoặc users.avatar_url
   (việc lưu URL này thuộc trách nhiệm của Event Service / Identity Service tương ứng,
   không phải File Service)
```

## Validate trước khi cấp presigned URL

- **Loại file**: whitelist theo `content_type` (vd: `image/jpeg`, `image/png`, `image/webp` cho banner/avatar) — từ chối loại file không nằm trong whitelist.
- **Kích thước**: giới hạn tối đa (vd: 5MB cho avatar, 10MB cho banner sự kiện) — nếu object storage hỗ trợ, nhúng điều kiện kích thước vào chính sách của presigned URL (policy) để storage tự chặn phía server, không chỉ dựa vào kiểm tra phía client.
- **Auth**: chỉ user đã đăng nhập (`authenticated`) mới gọi được `/files/presign` — không cấp presigned URL cho request ẩn danh, tránh bị lạm dụng làm nơi lưu trữ file tuỳ ý.

## Vì sao không upload qua backend

Nếu file đi qua backend (client → backend → object storage), backend phải giữ kết nối và xử lý băng thông cho toàn bộ nội dung file, tốn tài nguyên và không tận dụng được CDN gần người dùng. Với kiến trúc presigned URL, backend chỉ tham gia bước cấp quyền (rẻ, nhanh), còn việc truyền file nặng diễn ra trực tiếp giữa client và storage provider.

## Quan hệ với domain khác

- **event-catalog**: dùng presigned URL để upload banner sự kiện, lưu URL kết quả vào `events.banner_url`.
- **identity**: dùng presigned URL để upload avatar, lưu vào `users.avatar_url`.

## API liên quan

Xem [../../../api-docs/openapi/file-service.yaml](../../../api-docs/openapi/file-service.yaml).

## Phân theo phase

| Tính năng | Phase |
|---|---|
| `POST /files/presign` với validate loại file/kích thước | 2 (Must-have) |
