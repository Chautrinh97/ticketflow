# Tổng quan dự án

## Tên dự án

**TicketFlow** — nền tảng đặt vé sự kiện trực tuyến (concert, workshop, thể thao...).

## Bài toán

Xây dựng nền tảng cho phép **organizer** (đơn vị tổ chức) đăng sự kiện có số lượng vé giới hạn, và **người dùng** tìm kiếm, mua vé trực tuyến. Bài toán có ba đặc điểm bắt buộc hệ thống phải xử lý đúng:

1. **Số lượng vé/ghế hữu hạn, nhiều người có thể đặt cùng lúc** (kịch bản flash-sale) → hệ thống phải đảm bảo **không bán trùng vé** bằng transaction ACID kết hợp locking ở tầng dữ liệu (xem [docs/02-domains/booking/spec.md](../02-domains/booking/spec.md)).
2. **Mỗi loại sự kiện có tập thuộc tính khác nhau** — concert cần nghệ sĩ biểu diễn, workshop cần giảng viên/tài liệu, thể thao cần đội thi đấu → cần **schema linh hoạt** bên cạnh dữ liệu giao dịch chặt chẽ (xem [docs/03-data/mongodb-schema.md](../03-data/mongodb-schema.md)).
3. **Người dùng tìm sự kiện theo tên, địa điểm, có thể gõ sai chính tả** → cần **full-text + fuzzy search** (xem [docs/02-domains/event-catalog/spec.md](../02-domains/event-catalog/spec.md) và [docs/02-domains/search/spec.md](../02-domains/search/spec.md)).

## Mục tiêu dự án

Đây là một dự án portfolio thể hiện năng lực thiết kế hệ thống end-to-end: backend Go theo kiến trúc microservice, mô hình dữ liệu, bảo mật, tới frontend và vận hành hạ tầng (Docker/Kubernetes/CDN). Trọng tâm không chỉ là tính năng hoạt động được, mà là các quyết định thiết kế thể hiện tư duy hệ thống — transaction/locking đúng đắn, phân quyền theo cả role lẫn ownership, tách biệt trách nhiệm giữa các service, và khả năng vận hành production (observability, CI/CD, autoscaling).

## Đối tượng sử dụng tài liệu này

- Đọc [roles-permissions.md](roles-permissions.md) để hiểu ai được làm gì trong hệ thống.
- Đọc [tech-stack.md](tech-stack.md) để biết công nghệ dùng ở từng lớp.
- Đọc [../01-architecture/system-architecture.md](../01-architecture/system-architecture.md) để hiểu cách các service phối hợp.
- Đọc [../02-domains/](../02-domains/) để có spec chi tiết theo từng domain nghiệp vụ trước khi implement.
