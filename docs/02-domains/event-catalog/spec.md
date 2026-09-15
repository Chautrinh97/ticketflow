# Domain: Event Catalog

**Service sở hữu:** Event Service · **Database:** PostgreSQL (`events`, `ticket_types`) + MongoDB (`event_catalog`) · **Phase:** MVP (Phase 1) cho CRUD cơ bản, Phase 2 cho full-text/fuzzy search + validate `attributes`.

## Phạm vi & trách nhiệm

Quản lý vòng đời sự kiện: tạo/sửa/xoá, quản lý các loại vé (`ticket_types`) của mỗi sự kiện, lưu thuộc tính linh hoạt theo category, và phục vụ tìm kiếm/liệt kê sự kiện cho người dùng cuối. Event Service **không** xử lý đặt vé/tồn kho giao dịch (đó là trách nhiệm của Booking Service) — nó chỉ định nghĩa loại vé và `quota` ban đầu; `sold_count` được Booking Service cập nhật trong transaction đặt vé.

## Data model

- PostgreSQL `events`, `ticket_types` — xem [../../03-data/postgres-schema.md](../../03-data/postgres-schema.md#event-service--events-ticket_types).
- MongoDB `event_catalog` (thuộc tính linh hoạt theo category) — xem [../../03-data/mongodb-schema.md](../../03-data/mongodb-schema.md).

## Luồng nghiệp vụ

### Tạo & xuất bản sự kiện (organizer)

1. Organizer đăng nhập, tạo sự kiện (`status='draft'`) qua `POST /organizer/events` — ghi vào `events` (Postgres) và `event_catalog` (Mongo) trong cùng luồng xử lý của handler (nếu ghi Mongo thất bại, coi như tạo thất bại toàn bộ, không để lại `events` mồ côi không có catalog tương ứng).
2. Organizer thêm các `ticket_types` (giá, số lượng, tên loại vé) qua `POST /organizer/events/:id/ticket-types`.
3. Upload banner: gọi File Service lấy presigned URL (xem [../file-storage/spec.md](../file-storage/spec.md)) → upload thẳng lên object storage → gửi URL về lưu vào `events.banner_url`.
4. Organizer bấm "Xuất bản" → `status='published'`. Tại thời điểm này, Event Service tính lại `search_vector` (tsvector) từ `title + description` để phục vụ full-text search.
5. *(Phase 4)* Publish event `event.published` để Search Service đồng bộ sang Elasticsearch — xem [../search/spec.md](../search/spec.md).

Sự kiện ở trạng thái `draft` **không** hiển thị trong `GET /events` công khai — chỉ organizer sở hữu (hoặc `super_admin`) xem được qua endpoint quản lý riêng.

### Validate `attributes` theo category

Khi ghi `event_catalog.attributes`, Event Service validate shape tối thiểu theo `category` (vd: `concert` nên có `artists` là mảng không rỗng) trước khi chấp nhận — tránh catalog chứa dữ liệu rác không dùng được ở frontend. Danh sách category và trường gợi ý xem [../../03-data/mongodb-schema.md](../../03-data/mongodb-schema.md#attributes-theo-từng-category).

### Sửa / xoá sự kiện

`PATCH/DELETE /organizer/events/:id` — chỉ organizer sở hữu (`event.organizer_id == current_user.id`) hoặc `super_admin` (xem [../../04-security/authorization.md](../../04-security/authorization.md)). Xoá sự kiện đã có đơn hàng `paid` liên kết nên bị chặn ở tầng nghiệp vụ (trả lỗi rõ ràng) thay vì xoá cứng gây mất dữ liệu giao dịch của Booking Service — cần hủy toàn bộ sự kiện thì chuyển `status='cancelled'` thay vì `DELETE` vật lý.

### Tìm kiếm & liệt kê

| Loại | Cơ chế | Khi dùng |
|---|---|---|
| Normal | `WHERE city = ? AND category = ? AND start_time > ?` với index B-Tree | Lọc theo tiêu chí rõ ràng |
| Full-text | Postgres `tsvector`/`tsquery` trên `title + description` | Tìm theo từ khoá (Phase 2) |
| Fuzzy | Extension `pg_trgm` (`similarity()`) | Chấp nhận gõ sai chính tả (Phase 2) |
| Elasticsearch | *(nice-to-have)* — xem [../search/spec.md](../search/spec.md) | Khi cần scale tìm kiếm độc lập với DB giao dịch (Phase 4) |

`GET /events/search?q=` kết hợp full-text (ưu tiên) và fallback fuzzy nếu full-text không ra kết quả phù hợp.

## Quan hệ với domain khác

- **identity**: `events.organizer_id` tham chiếu `users.id`, chỉ chấp nhận user có `role='organizer'`.
- **booking**: `ticket_types.quota`/`sold_count` là nguồn dữ liệu Booking Service dùng để kiểm tra tồn kho khi đặt vé — Event Service không tự trừ `sold_count`.
- **file-storage**: banner sự kiện lấy URL qua File Service.
- **search** *(Phase 4)*: publish `event.published` khi xuất bản.

## API liên quan

Xem [../../../api-docs/openapi/event-service.yaml](../../../api-docs/openapi/event-service.yaml) — bao gồm cả `GET /organizer/events` (danh sách sự kiện của chính organizer, phục vụ màn hình [docs/06-frontend/screens/organizer/event-list.md](../../06-frontend/screens/organizer/event-list.md)) và `GET /organizer/events/{id}` (chi tiết 1 sự kiện theo id, phục vụ [docs/06-frontend/screens/organizer/event-manage.md](../../06-frontend/screens/organizer/event-manage.md) — ownership check giống hệt `PATCH/DELETE /organizer/events/{id}`).

## Phân theo phase

| Tính năng | Phase |
|---|---|
| CRUD sự kiện, ticket_types, `GET /events`, `GET /events/:slug` | 1 (MVP) |
| Full-text/fuzzy search, validate `attributes` theo category, cache Redis | 2 (Must-have) |
| Đồng bộ Elasticsearch | 4 (Nice-to-have) |
