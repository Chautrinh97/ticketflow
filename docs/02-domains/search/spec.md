# Domain: Search *(nice-to-have — Phase 4)*

**Service sở hữu:** Search Service · **Database:** Elasticsearch · **Phase:** 4 (Nice-to-have). Ở Phase 1-2, tìm kiếm nằm trong [event-catalog](../event-catalog/spec.md) dùng trực tiếp Postgres — xem mục "Tìm kiếm & liệt kê" tại đó.

## Phạm vi & trách nhiệm

Cung cấp tìm kiếm sự kiện có khả năng mở rộng độc lập với database giao dịch: full-text search chịu lỗi chính tả (fuzzy), autocomplete, xếp hạng kết quả theo độ liên quan. Tách thành service riêng khi nhu cầu tìm kiếm vượt quá khả năng đáp ứng hiệu quả của `tsvector`/`pg_trgm` trên Postgres, hoặc khi muốn scale tìm kiếm độc lập không ảnh hưởng tải của database giao dịch.

## Đồng bộ dữ liệu

Search Service không phải nguồn sự thật — nó chỉ giữ một **bản sao được đánh index** của dữ liệu sự kiện đã `published`, đồng bộ từ Event Service:

- Nhận event `event.published` (và các event cập nhật/huỷ sự kiện tương ứng) từ Event Service qua message queue — xem [../../01-architecture/event-driven-design.md](../../01-architecture/event-driven-design.md).
- Với dữ liệu thay đổi ngoài luồng event (backfill ban đầu, sửa lỗi lệch dữ liệu), chạy đồng bộ theo lịch (mỗi 5 phút) hoặc CDC (Change Data Capture) từ Postgres/Mongo.

## Truy vấn

- `fuzziness: AUTO` cho phép sai lệch chính tả theo độ dài từ khoá.
- Edge-ngram tokenizer cho autocomplete (gợi ý khi người dùng đang gõ).
- Xếp hạng theo `_score` mặc định của Elasticsearch, có thể boost thêm theo mức độ phổ biến (số vé đã bán) hoặc độ gần về thời gian diễn ra.

`GET /events/search?q=` ở API Gateway route sang Search Service thay vì Event Service khi Search Service đã sẵn sàng — về phía client, hợp đồng API (path, query param, response shape) giữ nguyên như định nghĩa tại [../../../api-docs/openapi/event-service.yaml](../../../api-docs/openapi/event-service.yaml), tránh phải đổi frontend khi chuyển đổi backend tìm kiếm.

## Quan hệ với domain khác

Consume `event.published` từ **event-catalog**. Không publish event nào cho domain khác — chỉ phục vụ query đọc.

## Phân theo phase

| Tính năng | Phase |
|---|---|
| Tách Search Service, index Elasticsearch, đồng bộ qua event, fuzzy + autocomplete | 4 (Nice-to-have) |
