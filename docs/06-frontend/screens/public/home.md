# Trang chủ

**Route:** `/` · **Vai trò:** public · **Render mode:** ISR · **Phase:** 1 (MVP)

**Liên quan:** [docs/02-domains/event-catalog/spec.md](../../02-domains/event-catalog/spec.md) · [api-docs/openapi/event-service.yaml](../../../api-docs/openapi/event-service.yaml)

## Layout tổng quan

```
┌─────────────────────────────────────────┐
│ Header (variant public)                  │
├─────────────────────────────────────────┤
│ [Hero] Banner lớn + SearchBar nổi bật     │
├─────────────────────────────────────────┤
│ [Section] Danh mục nổi bật (chip)         │
├─────────────────────────────────────────┤
│ [Section] Sự kiện sắp diễn ra (lưới card) │
│   [Nút "Xem thêm"]                        │
├─────────────────────────────────────────┤
│ Footer                                    │
└─────────────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Header, Footer, SearchBar | Component dùng chung — [components/layout.md](../../components/layout.md), [components/buttons-inputs.md](../../components/buttons-inputs.md) |
| Hero, Danh mục nổi bật, Danh sách sự kiện | Riêng màn hình này |
| Mỗi item trong lưới sự kiện | Component dùng chung — [`EventCard`](../../components/data-display.md#eventcard) |

## Chi tiết section riêng

- **Hero**: nền `bg-gray-50` cao khoảng 320px, tiêu đề `Display` ("Tìm sự kiện tiếp theo của bạn"), `SearchBar` (variant đầy đủ) căn giữa bên dưới tiêu đề.
- **Danh mục nổi bật**: hàng chip ngang (`concert`/`workshop`/`sport`, có thể cuộn ngang dưới `md`), mỗi chip `rounded-full border px-4 py-2 text-sm`, click điều hướng sang event-search.md kèm `?category=`.
- **Sự kiện sắp diễn ra**: tiêu đề section (`H2`) "Sự kiện sắp diễn ra" + lưới `EventCard` (1 cột dưới `md`, 2 cột `md`, 3 cột `lg` — theo breakpoint ở [design-system.md](../../design-system.md#breakpoints)), lấy từ `GET /events?from=<hôm nay>` sắp theo `start_time` tăng dần.

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Nhập từ khoá vào `SearchBar`, nhấn Enter | Điều hướng sang event-search.md kèm `?q=` |
| Click 1 chip danh mục | Điều hướng sang event-search.md kèm `?category=` |
| Click 1 `EventCard` | Điều hướng sang event-detail.md tương ứng |
| Bấm "Xem thêm" | Tải thêm trang tiếp theo của `GET /events`, nối vào cuối lưới hiện tại (quy ước "Xem thêm" — xem [interaction-patterns.md](../../interaction-patterns.md#pagination)) |

## Trạng thái đặc biệt

- **Loading lần đầu**: `Skeleton` dạng lưới `EventCard` (xem [components/feedback.md](../../components/feedback.md#loadingspinner--skeleton)).
- **Không có sự kiện nào sắp diễn ra**: `EmptyState` với text "Chưa có sự kiện nào sắp diễn ra, quay lại sau nhé!" (không có CTA vì đây là trang public, không phải trang organizer).

## Responsive

Hero thu chiều cao còn ~220px dưới `md`; lưới sự kiện chuyển 1 cột dưới `md`.
