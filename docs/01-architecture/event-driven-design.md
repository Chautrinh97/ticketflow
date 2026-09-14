# Event-driven & Distributed system *(nice-to-have — Phase 4)*

Ở MVP/Must-have (Phase 1-2), giao tiếp giữa Booking/Payment/Notification có thể triển khai bằng gọi gRPC trực tiếp theo kiểu request/response đồng bộ hoặc hàng đợi đơn giản trong tiến trình, miễn là hành vi nghiệp vụ (đặc biệt là compensating transaction khi thanh toán thất bại/hết hạn) đúng như mô tả. Tài liệu này mô tả kiến trúc **event-driven đầy đủ** dự kiến triển khai ở Phase 4.

## Message broker

RabbitMQ (dễ vận hành cho dự án cá nhân) hoặc Kafka (nếu muốn thể hiện khả năng chịu tải cao hơn). Chọn một, không dùng song song hai broker.

## Danh sách event chính

| Event | Publisher | Consumer(s) | Ý nghĩa |
|---|---|---|---|
| `order.created` | Booking Service | Payment Service | Đơn hàng vừa được tạo (trạng thái `pending`), tồn kho vé đã tạm trừ, chờ thanh toán |
| `payment.success` | Payment Service | Booking Service, Notification Service | Thanh toán thành công, xác nhận qua webhook |
| `payment.failed` | Payment Service | Booking Service, Notification Service | Thanh toán thất bại hoặc bị huỷ |
| `ticket.issued` | Booking Service | Notification Service | Vé (kèm mã QR) đã được sinh sau khi thanh toán thành công |
| `order.cancelled` | Booking Service | Notification Service | Đơn hàng bị huỷ (do thanh toán thất bại hoặc hết hạn giữ chỗ) |
| `event.published` | Event Service | Search Service | Sự kiện được organizer xuất bản, cần đồng bộ sang index tìm kiếm |

## Saga pattern (choreography)

Mỗi service tự lắng nghe event liên quan và thực hiện hành động bù trừ (compensating action) khi có lỗi, thay vì có một service trung tâm điều phối toàn bộ giao dịch phân tán (orchestration). Ví dụ: khi Booking Service nhận `payment.failed`, nó tự hoàn `sold_count` của `ticket_types` và chuyển `orders.status` sang `cancelled` — không cần một "saga coordinator" ra lệnh.

**Trade-off cần lưu ý khi trình bày thiết kế:** choreography giảm coupling (các service không biết về nhau, chỉ biết về event) nhưng làm luồng nghiệp vụ khó theo dõi tổng thể hơn (logic trải rộng ở nhiều service thay vì tập trung một chỗ) và khó debug hơn khi có lỗi giữa chừng — cần dựa vào tracing (xem [../05-infra-devops/observability.md](../05-infra-devops/observability.md)) để theo dõi một giao dịch xuyên suốt nhiều service.

## Idempotency

Mọi consumer phải xử lý idempotent — cùng một event có thể được deliver lại (do retry của broker, do consumer crash giữa chừng). Ví dụ: Notification Service không được gửi trùng email nếu nhận lại `ticket.issued` cho cùng một `order_id`; Booking Service không được trừ tồn kho hai lần nếu nhận lại `order.created` (không áp dụng — `order.created` do chính Booking Service publish, nhưng nguyên tắc idempotent áp dụng cho mọi consumer khác tương tự).
