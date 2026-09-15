# Scripts

**Trạng thái:** Phase 1 (MVP).

| Script | Mục đích |
|---|---|
| `gen-proto.sh` | Regenerate Go gRPC code từ `src/proto/*.proto` (yêu cầu `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`) |
| `seed-dev-users.sh` | Seed tài khoản `super_admin`/`organizer` trực tiếp vào Postgres cho dev (Phase 1 chưa có luồng duyệt organizer) |
| `smoke-test.sh` | Kiểm thử end-to-end qua API Gateway: login → tạo/publish sự kiện → thêm loại vé → đặt vé nhiều loại vé → mock thanh toán → xác nhận đã paid + vé đã phát hành |

Chạy migration: xem `deployments/docker-compose.yaml` (`migrate-*` services, dùng image `migrate/migrate` chính thức, chạy tự động khi `docker compose up`).
