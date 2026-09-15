# Proto

**Trạng thái:** Phase 1 — `identity.proto`, `event.proto`, `booking.proto` đã định nghĩa và generate.

Định nghĩa gRPC dùng cho giao tiếp **nội bộ service-to-service** (không phải hợp đồng public — hợp đồng public là `api-docs/openapi/*.yaml`, do chính mỗi service phục vụ qua REST). Mỗi service vẫn chạy REST server riêng khớp OpenAPI của nó; gRPC chỉ dùng cho các lời gọi nội bộ mà REST không phù hợp (vd: kiểm tra session còn hiệu lực, xác nhận thanh toán xuyên service).

Không có `payment.proto` — trong Phase 1 không service nào gọi payment-service qua gRPC (payment-service chỉ là caller vào booking-service và là REST responder cho gateway/webhook).

## Quy ước

- Một file `.proto` theo tên service, `option go_package = "ticketflow/proto/<name>pb;<name>pb"`.
- Message request/response phản ánh đúng field trong OpenAPI tương ứng khi khái niệm trùng nhau (vd: `Order`, `Event`), nhưng chỉ chứa field mà bên gọi nội bộ thực sự cần — không copy nguyên schema OpenAPI.
- Regenerate code: `scripts/gen-proto.sh` (yêu cầu `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc` trong `$PATH`). Output nằm ở `<name>pb/` (vd: `identitypb/`), tách thư mục theo package vì mỗi proto khai báo package Go khác nhau.
