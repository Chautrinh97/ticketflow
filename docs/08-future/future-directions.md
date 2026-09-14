# Hướng mở rộng tương lai

Các hướng mở rộng dưới đây nằm ngoài phạm vi 4 phase hiện tại ([../07-roadmap/](../07-roadmap/)) — chỉ ghi nhận làm định hướng dài hạn, chưa có kế hoạch triển khai cụ thể.

- **Đa ngôn ngữ, đa tiền tệ.**
- **"Phòng chờ ảo" (virtual waiting room)** khi flash-sale cực lớn, thay vì chỉ dựa vào rate-limit — xếp hàng người dùng trước khi cho vào trang checkout, giảm tải trực tiếp lên Booking Service tại thời điểm mở bán.
- **Dynamic pricing** theo nhu cầu — giá vé tăng khi vé sắp hết, cần thiết kế lại cách hiển thị giá ở Event Service và cách tính `unit_price` khi tạo đơn ở Booking Service (hiện tại `unit_price` chốt tại thời điểm đặt, xem [../03-data/postgres-schema.md](../03-data/postgres-schema.md)).
- **Ứng dụng mobile** (React Native) dùng chung API Gateway — không cần thêm backend riêng nếu REST API đã đủ tổng quát.
- **Multi-tenant** cho các organizer lớn cần thương hiệu/subdomain riêng.
- **Engine hoàn tiền/refund policy** có thể cấu hình theo từng sự kiện — hiện tại luồng huỷ vé ([../02-domains/booking/spec.md](../02-domains/booking/spec.md)) dùng một chính sách chung, chưa cho phép organizer tự định nghĩa điều kiện hoàn tiền riêng.
