# Danh sách / Tìm kiếm sự kiện

**Route:** `/events` (query `?q=`, `?category=`, `?city=`, `?from=`) · **Vai trò:** public · **Render mode:** SSR cho lần tải đầu theo query trên URL (để link chia sẻ/SEO đúng kết quả) + CSR khi người dùng đổi filter tiếp theo · **Phase:** 1 (filter cơ bản theo category/city/date) → 2 (full-text/fuzzy qua `q`)

**Liên quan:** [docs/02-domains/event-catalog/spec.md](../../02-domains/event-catalog/spec.md) (mục "Tìm kiếm & liệt kê") · [api-docs/openapi/event-service.yaml](../../../api-docs/openapi/event-service.yaml)

## Layout tổng quan

```
┌─────────────────────────────────────────┐
│ Header (variant public)                  │
├───────────┬─────────────────────────────┤
│ [Filter   │ [Thanh kết quả] "128 sự kiện"│
│  Panel]   │ [Lưới EventCard]             │
│  - category│                              │
│  - city    │  [Nút "Xem thêm"]            │
│  - ngày    │                              │
├───────────┴─────────────────────────────┤
│ Footer                                    │
└─────────────────────────────────────────┘
```

| Section | Loại |
|---|---|
| Header, Footer | Component dùng chung — [components/layout.md](../../components/layout.md) |
| `EventCard` trong lưới | Component dùng chung — [components/data-display.md](../../components/data-display.md#eventcard) |
| `Select` (category/city), `DatePicker` trong Filter Panel | Component dùng chung — [components/buttons-inputs.md](../../components/buttons-inputs.md) |
| Filter Panel (bố cục), Thanh kết quả | Riêng màn hình này |

## Chi tiết section riêng

- **Filter Panel**: cột trái rộng `w-64`, cố định (`sticky top-20`) trên `lg`; dưới `lg` thu gọn thành nút "Bộ lọc" mở `Modal` chứa cùng nội dung filter. Gồm: `Select` category, `Select` city, `DatePicker` khoảng ngày. Nút "Xoá bộ lọc" (`Button` variant `ghost`, chỉ hiện khi có ít nhất 1 filter đang áp dụng).
- **Thanh kết quả**: dòng text `Body small` hiển thị số lượng kết quả (vd "128 sự kiện") + `Select` sắp xếp (mặc định "Gần diễn ra nhất", tuỳ chọn "Giá thấp đến cao"/"Giá cao đến thấp").
- **Lưới `EventCard`**: giống bố cục ở home.md (1/2/3 cột theo breakpoint).

## Hành vi tương tác

| Hành động | Phản hồi hệ thống |
|---|---|
| Đổi 1 filter (category/city/ngày/sắp xếp) | Cập nhật query param trên URL, gọi lại `GET /events` (filter thường) hoặc `GET /events/search?q=` (nếu có `q`) — không reload trang, có `Skeleton` trong lúc chờ |
| Bấm "Xoá bộ lọc" | Xoá toàn bộ query param filter (giữ `q` nếu có), tải lại danh sách |
| Nhập `q` (đến từ SearchBar ở Header hoặc trang chủ) (Phase 2) | Ưu tiên gọi `GET /events/search?q=` — full-text trước, fallback fuzzy nếu không có kết quả phù hợp (theo domain spec) |
| Click `EventCard` | Điều hướng sang event-detail.md |
| Bấm "Xem thêm" | Tải thêm trang tiếp theo, nối vào cuối lưới (quy ước "Xem thêm") |

## Trạng thái đặc biệt

- **Không có kết quả phù hợp**: `EmptyState` — "Không tìm thấy sự kiện phù hợp với '<q>'. Thử từ khoá khác hoặc bỏ bớt bộ lọc."
- **Loading khi đổi filter**: giữ lưới cũ hiển thị mờ (`opacity-60`) thay vì xoá trắng, theo quy tắc loading tại [interaction-patterns.md](../../interaction-patterns.md#loading).

## Responsive

Dưới `lg`: Filter Panel ẩn thành nút "Bộ lọc" (badge số lượng filter đang áp dụng) mở `Modal`; lưới 1 cột dưới `md`.
