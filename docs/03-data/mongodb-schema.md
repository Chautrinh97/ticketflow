# MongoDB — catalog linh hoạt

Thuộc tính sự kiện thay đổi theo `category` (concert cần nghệ sĩ, workshop cần giảng viên, thể thao cần đội thi đấu) không phù hợp để mô hình hoá bằng cột cố định trong PostgreSQL. Phần này lưu ở MongoDB, tham chiếu ngược về `events.id` bên Postgres qua `event_id`.

## Collection `event_catalog`

```json
{
  "event_id": "uuid-tham-chiếu-tới-postgres",
  "category": "concert",
  "attributes": {
    "artists": ["Tên nghệ sĩ"],
    "stage_layout": "GA + VIP",
    "age_restriction": 16
  },
  "gallery_images": ["https://cdn.../1.jpg", "https://cdn.../2.jpg"],
  "faq": [
    { "question": "Có được đổi vé không?", "answer": "..." }
  ],
  "tags": ["nhạc trẻ", "v-pop", "hà nội"]
}
```

## `attributes` theo từng category

`attributes` là object tự do, không có schema cố định ở tầng database, nhưng tầng application (Event Service) validate theo category trước khi ghi:

| Category | Trường gợi ý trong `attributes` |
|---|---|
| `concert` | `artists` (mảng nghệ sĩ), `stage_layout`, `age_restriction` |
| `workshop` | `instructors` (mảng giảng viên), `materials` (tài liệu đi kèm), `max_participants` |
| `sport` | `teams` (mảng đội thi đấu), `match_format`, `venue_capacity` |

Khi thêm category mới, cập nhật bảng này và logic validate trong `src/services/event-service` — không để `attributes` nhận dữ liệu tuỳ ý không qua kiểm tra tối thiểu về shape theo category.

## Quan hệ với PostgreSQL

`event_catalog.event_id` **không** phải khoá ngoại thật (khác database) — tính toàn vẹn được đảm bảo ở tầng application: Event Service tạo document `event_catalog` ngay sau khi insert `events` thành công, trong cùng một luồng xử lý của handler (không cần 2-phase commit; nếu bước ghi Mongo thất bại, coi như tạo sự kiện thất bại và rollback phía Postgres). Chi tiết luồng tạo sự kiện xem [../02-domains/event-catalog/spec.md](../02-domains/event-catalog/spec.md).

## Index

- Index theo `event_id` (unique) — tra cứu catalog theo sự kiện.
- Index theo `tags` — phục vụ lọc theo tag ở Event Service (lọc cơ bản, khác với full-text search chạy trên Postgres/Elasticsearch).
