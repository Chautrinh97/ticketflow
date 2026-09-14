# Kiến trúc hệ thống

## Pattern

Microservice, mỗi service sở hữu database riêng (**database-per-service**). Giao tiếp nội bộ giữa các service qua **gRPC**; giao tiếp với client (web/mobile) qua **REST** thông qua API Gateway.

## Danh sách service

| Service | Trách nhiệm chính | Giao tiếp | Database chính |
|---|---|---|---|
| API Gateway | Routing, xác thực JWT, rate limit, CORS, tổng hợp response | REST (inbound) ⇄ gRPC (outbound) | Redis (rate-limit state) |
| Identity Service | Đăng ký/đăng nhập, phát JWT, quản lý user & role | REST + gRPC | PostgreSQL |
| Event Service | CRUD sự kiện, loại vé, tìm kiếm cơ bản | REST + gRPC | PostgreSQL (+ MongoDB cho thuộc tính linh hoạt) |
| Booking Service | Tạo đơn, giữ chỗ, đảm bảo ACID khi trừ tồn kho vé | REST + gRPC | PostgreSQL |
| Payment Service | Tích hợp cổng thanh toán, xử lý webhook | REST + gRPC | PostgreSQL |
| Notification Service | Gửi email/push, lưu thông báo in-app | Consume message queue | PostgreSQL |
| File Service | Sinh presigned URL upload lên object storage | REST | — (stateless, gọi S3-compatible API) |
| Search Service *(nice-to-have)* | Đồng bộ & truy vấn Elasticsearch | REST + gRPC | Elasticsearch |
| Analytics Service *(nice-to-have)* | Tổng hợp số liệu, dashboard | Consume message queue | PostgreSQL / ClickHouse |

Ở MVP, **Search** nằm trong Event Service (dùng `tsvector`/`pg_trgm` của Postgres); tách thành service riêng + Elasticsearch ở Phase 4 — xem [../02-domains/search/spec.md](../02-domains/search/spec.md).

## Nguyên tắc giao tiếp giữa service

- **Client → hệ thống**: luôn đi qua API Gateway bằng REST. Không service nghiệp vụ nào lộ endpoint trực tiếp ra ngoài.
- **Service → service (đồng bộ)**: gRPC, có interceptor xử lý auth (truyền `trace_id`/user context), logging, tracing.
- **Service → service (bất đồng bộ)**: qua message queue theo mô hình event-driven, chi tiết tại [event-driven-design.md](event-driven-design.md). Notification Service **chỉ** là consumer — không service nào gọi trực tiếp vào nó.
- **Database-per-service**: một service không được truy vấn trực tiếp database của service khác. Muốn lấy dữ liệu thuộc domain khác, gọi qua gRPC (đồng bộ) hoặc lắng nghe event (bất đồng bộ) — không join cross-database.

## Sơ đồ luồng request điển hình (đặt vé)

```
Client
  │  REST
  ▼
API Gateway  ──(xác thực JWT, rate-limit)──┐
  │ gRPC                                    │
  ▼                                         │
Booking Service ──(SELECT...FOR UPDATE, transaction)──▶ PostgreSQL (booking DB)
  │ publish "order.created"
  ▼
Message Queue
  │
  ├──▶ Payment Service ──(khởi tạo phiên thanh toán)──▶ Payment Gateway
  │        │ webhook callback
  │        └─ publish "payment.success" / "payment.failed"
  │
  └──▶ Notification Service ──(consume các event)──▶ Email/in-app/push
```

Chi tiết từng bước xem [../02-domains/booking/spec.md](../02-domains/booking/spec.md).

## API Gateway — vai trò cụ thể

- Xác thực JWT (access token), từ chối request không hợp lệ trước khi tới service nghiệp vụ.
- Rate limit theo IP/user (token bucket, state lưu Redis) — mức giới hạn khác nhau theo route, đặc biệt nghiêm ngặt với endpoint đặt vé (xem [../04-security/rate-limiting.md](../04-security/rate-limiting.md)).
- CORS: chỉ whitelist đúng domain frontend.
- Routing REST → gRPC tới đúng service, tổng hợp response khi cần gộp dữ liệu từ nhiều service cho một API công khai.

## Vì sao chọn database-per-service thay vì database chung

Cô lập lỗi và thay đổi schema theo từng domain — một migration lỗi ở Event Service không ảnh hưởng Booking Service; mỗi service có thể chọn loại DB phù hợp nhất với đặc thù dữ liệu (Postgres cho giao dịch, Mongo cho catalog linh hoạt). Đánh đổi là không thể JOIN trực tiếp qua domain, phải chấp nhận eventual consistency ở một số chỗ (xử lý qua event, xem [event-driven-design.md](event-driven-design.md)) — đây là trade-off cần giải thích rõ khi trình bày thiết kế.
