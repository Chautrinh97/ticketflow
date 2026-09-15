# Cấu trúc thư mục repo

```
.
├── docs/                     # Tài liệu spec — nguồn sự thật cho thiết kế
├── api-docs/                 # Hợp đồng API (OpenAPI 3.0) cho từng service
├── src/
│   ├── services/              # Các microservice Go
│   │   ├── api-gateway/
│   │   ├── identity-service/
│   │   ├── event-service/
│   │   ├── booking-service/
│   │   ├── payment-service/
│   │   ├── notification-service/
│   │   └── file-service/
│   ├── frontend/
│   │   └── nextjs-app/        # Ứng dụng Next.js
│   ├── proto/                 # Định nghĩa gRPC dùng chung giữa các service
│   ├── pkg/                   # Code Go dùng chung giữa nhiều service (xem mục riêng bên dưới)
│   └── go.mod                 # 1 Go module duy nhất cho toàn bộ services/ + proto/ + pkg/
├── deployments/
│   ├── docker-compose.yaml    # Chạy toàn bộ hệ thống ở local
│   ├── k8s/
│   │   ├── base/               # Manifest Kubernetes dùng chung
│   │   └── overlays/{dev,prod} # Override theo môi trường (kustomize)
│   └── helm/                  # Helm chart
├── scripts/                   # Script tiện ích cho dev/CI
├── .claude/                   # Cấu hình Claude Code cho repo này
├── AGENTS.md                  # Quy ước chung cho AI coding agent
└── README.md
```

## Quy ước layout bên trong một service Go (`src/services/<name>/`)

```
<name>/
├── cmd/                # Entrypoint (main.go)
├── internal/
│   ├── handler/
│   │   ├── http/        # REST handler (Gin) — khớp api-docs/openapi/<name>-service.yaml
│   │   └── grpc/        # gRPC server — khớp src/proto/<name>.proto (RPC nội bộ, không phải hợp đồng public)
│   ├── service/         # Business logic
│   ├── repository/      # Truy cập database
│   └── model/           # Struct dữ liệu nội bộ
├── migrations/          # golang-migrate: *.up.sql / *.down.sql
└── Dockerfile
```

- `handler` không chứa business logic — mọi quyết định nghiệp vụ nằm ở `service`. Tách `handler/http` và `handler/grpc` vì phần lớn service có cả hai: REST cho client (qua API Gateway) và gRPC cho lời gọi nội bộ từ service khác.
- `repository` là lớp duy nhất chạm database — `service` không tự viết SQL/query trực tiếp (ngoại lệ: `booking-service` dùng raw SQL/pgx thay vì ORM ngay trong `repository`, để kiểm soát `SELECT ... FOR UPDATE` — vẫn là lớp duy nhất chạm DB, chỉ khác công cụ).
- Mỗi service có `Dockerfile` multi-stage riêng (build bằng Go builder image → copy binary vào image `alpine`); build context là `src/` (không phải thư mục service) vì mọi service dùng chung 1 `go.mod`.

## `src/pkg/` — code Go dùng chung giữa các service

Vì toàn bộ service nằm trong **1 Go module duy nhất** (`src/go.mod`), package `internal/` của một service không thể được import bởi service khác (giới hạn ngôn ngữ Go) — `src/pkg/` là nơi duy nhất hợp lệ cho code thực sự cần dùng chung, và chỉ chứa những gì đang thực sự được nhiều service dùng (không thêm trước "phòng khi cần"):

| Package | Dùng cho |
|---|---|
| `pkg/authclaims` | Struct JWT claim (`sub`,`role`,`jti`,`exp`) + `Sign`/`Parse` — identity-service phát token, mọi service khác + API Gateway tự verify local |
| `pkg/httpauth` | Middleware Gin `RequireAuth`/`RequireRole`/`RequireOwnership` (chữ ký khớp docs/04-security/authorization.md) |
| `pkg/apperr` | Kiểu lỗi `{code,message}` dùng chung + map sang HTTP status / gRPC code |
| `pkg/pagination` | Envelope `{items,total,page,page_size}` dùng chung cho mọi endpoint list |
| `pkg/grpcinterceptor` | Interceptor logging/recovery/trace-id cho gRPC client và server nội bộ |

Không đặt business logic của riêng 1 domain vào `pkg/` — chỉ hạ tầng/tiện ích ngang hàng thực sự trung lập giữa các service.

## Vì sao gộp source code dưới `src/`

Toàn bộ `services/`, `frontend/`, `proto/` nằm dưới một thư mục `src/` duy nhất ở root, tách biệt rõ ràng với `docs/`, `api-docs/`, `deployments/` — giữ root repo gọn, dễ phân biệt "thứ được biên dịch/chạy" (`src/`) với "tài liệu và cấu hình vận hành" (mọi thư mục còn lại).
