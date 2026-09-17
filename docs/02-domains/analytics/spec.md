# Domain: Analytics

**Service sở hữu:** Phase 2 — tính toán nằm trong Booking/Event Service (query trực tiếp); Phase 4 — tách **Analytics Service** riêng · **Database:** PostgreSQL (Phase 2), ClickHouse (Phase 4) · **Phase:** thống kê organizer cơ bản là **must-have** (Phase 2), Analytics Service độc lập là **nice-to-have** (Phase 4).

## Phạm vi & trách nhiệm

Cung cấp số liệu thống kê cho hai đối tượng:

- **Organizer**: doanh thu, tỉ lệ bán vé theo thời gian thực cho các sự kiện **của chính mình**.
- **Super Admin**: thống kê toàn hệ thống (tổng doanh thu, số đơn hàng, tăng trưởng...).

## Giai đoạn Phase 2 — thống kê tính trực tiếp (must-have)

Không cần service riêng — thống kê được tính bằng query tổng hợp (`SUM`, `COUNT`, `GROUP BY`) trực tiếp trên dữ liệu `orders`/`order_items`/`tickets` (thuộc Booking Service) và `events`/`ticket_types` (thuộc Event Service). Vì đây là truy vấn xuyên domain, nơi thực thi hợp lý là ở tầng gọi tổng hợp qua gRPC (Booking Service expose một RPC nội bộ trả về số liệu thô cho Event Service hoặc cho API Gateway tổng hợp), **không** cho một service query thẳng vào database của service khác.

- `GET /organizer/stats` (Booking Service — xem [../../../api-docs/openapi/booking-service.yaml](../../../api-docs/openapi/booking-service.yaml)) — tổng quan doanh thu/tỉ lệ bán vé của organizer hiện tại trên mọi sự kiện.
- `GET /organizer/events/:id/stats` (Booking Service) — doanh thu, số vé đã bán/còn lại theo từng `ticket_type` của 1 sự kiện. Yêu cầu ownership check giống mọi endpoint organizer khác (xem [../../04-security/authorization.md](../../04-security/authorization.md)).
- `GET /admin/stats` (Booking Service) — chỉ `super_admin`, số liệu toàn hệ thống.
- **Báo cáo doanh thu ngày**: cron job 0h hằng ngày tổng hợp doanh thu theo organizer, lưu snapshot vào bảng `organizer_revenue_daily_snapshots` (schema: [../../03-data/postgres-schema.md](../../03-data/postgres-schema.md)) — xem [../../05-infra-devops/background-jobs.md](../../05-infra-devops/background-jobs.md). Mục đích snapshot: tránh phải quét lại toàn bộ `orders` mỗi lần muốn xem lịch sử doanh thu theo ngày.

## Giai đoạn Phase 4 — Analytics Service riêng (nice-to-have)

Khi khối lượng dữ liệu hoặc độ phức tạp truy vấn (phân tích xu hướng, so sánh nhiều chiều) vượt quá khả năng query trực tiếp hiệu quả trên Postgres giao dịch:

- Tách **Analytics Service**, consume event (`order.created`, `payment.success`, `ticket.issued`...) qua message queue để xây dựng bảng tổng hợp riêng, không đọc trực tiếp database giao dịch của Booking/Event Service.
- Lưu trữ ở **ClickHouse** — phù hợp cho truy vấn phân tích (OLAP) quy mô lớn, tách biệt khỏi database phục vụ giao dịch (OLTP).

## Quan hệ với domain khác

Đọc dữ liệu tổng hợp từ **booking** và **event-catalog** (Phase 2: qua gRPC nội bộ; Phase 4: qua consume event). Không ghi ngược lại dữ liệu vào các domain đó.

## Phân theo phase

| Tính năng | Phase |
|---|---|
| Thống kê doanh thu/tỉ lệ bán vé cho organizer, thống kê toàn hệ thống cho admin, báo cáo doanh thu ngày | 2 (Must-have) |
| Tách Analytics Service, ClickHouse, consume event | 4 (Nice-to-have) |
