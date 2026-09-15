# Chi tiết sự kiện

**Route:** `/events/[slug]` · **Vai trò:** public (đặt vé yêu cầu đăng nhập — xem mục Hành vi tương tác) · **Render mode:** SSR · **Phase:** 1 (MVP)

**Liên quan:** [docs/02-domains/event-catalog/spec.md](../../02-domains/event-catalog/spec.md) · [docs/03-data/mongodb-schema.md](../../03-data/mongodb-schema.md) · [api-docs/openapi/event-service.yaml](../../../api-docs/openapi/event-service.yaml)

## Layout tổng quan

```
┌─────────────────────────────────────────┐
│ Header (variant public)                  │
├─────────────────────────────────────────┤
│ [Banner sự kiện, full-width]              │
├───────────────────────────┬─────────────┤
│ [Nội dung chính]           │ [Panel vé]  │
│  - Tiêu đề + Badge category │  (sticky)   │
│  - Thời gian, địa điểm      │  - Danh sách│
│  - Mô tả                    │    TicketTy-│
│  - Thuộc tính theo category │    peRow    │
│  - FAQ                      │  - Nút      │
│                              │    "Đặt vé" │
├───────────────────────────┴─────────────┤
│ Footer                                    │
└─────────────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Header, Footer | Component dùng chung — [components/layout.md](../../components/layout.md) |
| Badge category, mỗi dòng vé | Component dùng chung — [`Badge`](../../components/data-display.md#badge--statustag), [`TicketTypeRow`](../../components/data-display.md#tickettyperow) |
| Nội dung chính, Panel vé (bố cục) | Riêng màn hình này |

## Chi tiết section riêng

- **Banner**: ảnh `events.banner_url`, tỉ lệ 21:9, full-width, phủ gradient tối nhẹ phía dưới nếu có overlay tiêu đề (tuỳ chọn thiết kế, không bắt buộc).
- **Nội dung chính** (cột trái, ~65% chiều rộng trên `lg`):
  - Tiêu đề (`H1`) + `Badge` category cạnh tiêu đề.
  - Dòng meta: icon lịch + `start_time`–`end_time` định dạng "20:00, Thứ Bảy 12/09/2026", icon ghim + `venue_name`, `address`, `city`.
  - Mô tả (`events.description`, giữ format xuống dòng).
  - **Thuộc tính theo category** (`event_catalog.attributes`): render khác nhau theo `category` — `concert` hiện danh sách nghệ sĩ dạng chip; `workshop` hiện danh sách giảng viên + tài liệu đính kèm (nếu có link); `sport` hiện danh sách đội thi đấu. Xem cấu trúc dữ liệu tại [docs/03-data/mongodb-schema.md](../../03-data/mongodb-schema.md#attributes-theo-từng-category).
  - **FAQ**: danh sách accordion (câu hỏi bấm mở ra câu trả lời), từ `event_catalog.faq`.
- **Panel vé** (cột phải, ~35%, `sticky top-20` trên `lg`): khối `rounded-lg border border-gray-200 p-6` — tiêu đề "Chọn vé", danh sách `TicketTypeRow` (chỉ hiển thị, chưa cho chọn số lượng ở màn hình này) → `Button` variant `primary` full-width "Đặt vé ngay".

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Bấm 1 câu hỏi trong FAQ | Toggle mở/đóng câu trả lời tương ứng (accordion, không ảnh hưởng câu khác) |
| Bấm "Đặt vé ngay" — đã đăng nhập | Điều hướng sang checkout.md |
| Bấm "Đặt vé ngay" — chưa đăng nhập | Điều hướng sang auth-login.md, sau khi đăng nhập thành công tự động quay lại checkout.md của đúng sự kiện này |
| Tất cả loại vé đã hết (`sold_count == quota` với mọi `ticket_type`) | Nút "Đặt vé ngay" chuyển `disabled`, label đổi thành "Đã hết vé" |
| Sự kiện `status != published` hoặc không tồn tại | Hiển thị màn hình lỗi 404 — xem [system/error-states.md](../system/error-states.md) |

## Trạng thái đặc biệt

Loading: `Skeleton` cho Banner + khối nội dung chính + Panel vé, giữ đúng tỉ lệ layout thật.

## Responsive

Dưới `lg`: Panel vé không còn `sticky`/cột riêng — chuyển xuống dưới Nội dung chính; đồng thời hiện 1 thanh CTA cố định đáy màn hình (`fixed bottom-0`) chứa giá vé thấp nhất + nút "Đặt vé ngay" để luôn thao tác được mà không cần cuộn lên Panel vé.
