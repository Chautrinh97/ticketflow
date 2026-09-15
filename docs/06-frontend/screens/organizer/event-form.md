# Tạo / Chỉnh sửa sự kiện

**Route:** `/organizer/events/new` (tạo mới), `/organizer/events/[id]/edit` (chỉnh sửa — Phase 2) · **Vai trò:** organizer (chỉnh sửa: chỉ chủ sở hữu — ownership check, xem [docs/04-security/authorization.md](../../04-security/authorization.md)) · **Render mode:** CSR · **Phase:** 1 (tạo mới, không có ảnh) → 2 (chỉnh sửa sự kiện đã tạo, ảnh, thuộc tính theo category)

**Liên quan:** [docs/02-domains/event-catalog/spec.md](../../02-domains/event-catalog/spec.md) · [docs/02-domains/file-storage/spec.md](../../02-domains/file-storage/spec.md) · [api-docs/openapi/event-service.yaml](../../../api-docs/openapi/event-service.yaml)

Cùng 1 form dùng cho cả tạo mới và chỉnh sửa (chỉnh sửa: Phase 2 — các field được điền sẵn từ dữ liệu hiện có). Dùng layout dashboard chung — xem [organizer/dashboard.md](dashboard.md#layout-tổng-quan).

## Layout tổng quan

```
┌───────┬─────────────────────────────────┐
│Sidebar│ Header (variant dashboard)        │
│       ├───────────────────────────────────┤
│       │ [Section] Thông tin cơ bản         │
│       │  - Tên, category, mô tả            │
│       │  - Thời gian, địa điểm             │
│       │  - Banner (FileUpload) (Phase 2)   │
│       ├───────────────────────────────────┤
│       │ [Section] Thuộc tính theo category │
│       ├───────────────────────────────────┤
│       │ [Section] Loại vé (bảng, thêm dòng)│
│       ├───────────────────────────────────┤
│       │ [Thanh dưới cùng] Lưu nháp | Xuất  │
│       │   bản                              │
└───────┴─────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Header, Sidebar, `Input`, `Select`, `DatePicker`, `FileUpload`, `Button` | Component dùng chung |
| Thông tin cơ bản, Thuộc tính theo category, Loại vé, Thanh dưới cùng (bố cục & nội dung riêng) | Riêng màn hình này |

## Chi tiết section riêng

- **Thông tin cơ bản**: `Input` tên sự kiện, `Select` category, `Input` variant `textarea` mô tả, `DatePicker` thời gian bắt đầu/kết thúc, `Input` địa điểm (`venue_name`, `address`, `city`), `FileUpload` banner (tỉ lệ 16:9) (Phase 2).
- **Thuộc tính theo category**: form động thay đổi theo `category` đã chọn ở trên (khớp `event_catalog.attributes`, xem [docs/03-data/mongodb-schema.md](../../03-data/mongodb-schema.md#attributes-theo-từng-category)) — `concert`: input thêm nhiều nghệ sĩ (dạng tag input); `workshop`: input thêm giảng viên + link tài liệu; `sport`: input thêm đội thi đấu. Section này ẩn hoàn toàn cho tới khi đã chọn `category`.
- **Loại vé**: bảng các dòng (tên, giá, số lượng), nút "+ Thêm loại vé" thêm 1 dòng trống; mỗi dòng có `IconButton` xoá. Bắt buộc tối thiểu 1 loại vé mới cho lưu được.
- **Thanh dưới cùng** (`sticky bottom-0 bg-white border-t px-6 py-4 flex justify-end gap-3`): `Button` `secondary` "Lưu nháp" (giữ `status=draft`), `Button` `primary` "Xuất bản" (chuyển `status=published`, chỉ hiện khi đang ở `draft` hoặc lần đầu tạo).

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Chọn/đổi ảnh banner | Upload qua luồng presigned URL (xem [components/buttons-inputs.md](../../components/buttons-inputs.md#fileupload--imageuploader)) |
| Bấm "+ Thêm loại vé" | Thêm 1 dòng trống vào bảng Loại vé, focus vào ô tên vừa thêm |
| Bấm "Lưu nháp" | Validate các field bắt buộc tối thiểu (tên, category, thời gian) → gọi `POST /organizer/events` (tạo mới) hoặc `PATCH /organizer/events/:id` (chỉnh sửa — Phase 2) + `POST .../ticket-types` cho từng loại vé mới → thành công: `Toast` `success` "Đã lưu", ở lại trang (chuyển route sang `/edit/:id` nếu vừa tạo mới) |
| Bấm "Xuất bản" | Validate đầy đủ (tối thiểu 1 loại vé; yêu cầu bắt buộc có banner sẽ áp dụng từ Phase 2 khi có hỗ trợ ảnh) → lưu như trên rồi gọi `POST /organizer/events/:id/publish` → thành công: `Toast` `success` "Đã xuất bản sự kiện", điều hướng sang event-manage.md |
| Rời trang khi có thay đổi chưa lưu | Mở `ConfirmDialog` "Rời khỏi mà không lưu?" (theo quy tắc chung ở [interaction-patterns.md](../../interaction-patterns.md#modal--dialog)) |

## Trạng thái đặc biệt

Lỗi validate: `InlineFormError` dưới từng field liên quan, không cho submit tới khi sửa đúng.

## Responsive

Toàn bộ section xếp dọc 1 cột ở mọi kích thước màn hình (form dài, không cần bố cục nhiều cột).
