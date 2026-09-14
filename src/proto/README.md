# Proto

**Trạng thái:** chưa có định nghĩa `.proto` — placeholder scaffold.

Định nghĩa gRPC dùng chung giữa các service (giao tiếp nội bộ service-to-service). Sẽ được bổ sung khi bắt đầu Phase 1 implementation, đồng bộ với các RPC method mô tả gián tiếp qua giao tiếp giữa các domain trong [../../docs/01-architecture/system-architecture.md](../../docs/01-architecture/system-architecture.md).

## Quy ước dự kiến

- Một file `.proto` theo tên service (vd: `identity.proto`, `event.proto`, `booking.proto`).
- Message request/response nên phản ánh đúng field đã định nghĩa trong `components.schemas` của OpenAPI tương ứng tại [../../api-docs/openapi/](../../api-docs/openapi/), tránh định nghĩa hai shape dữ liệu lệch nhau cho cùng một khái niệm (vd: `User`, `Event`, `Order`).
