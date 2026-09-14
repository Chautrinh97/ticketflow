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
│   └── proto/                 # Định nghĩa gRPC dùng chung giữa các service
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

Áp dụng khi bắt đầu implement (chưa tạo ở giai đoạn scaffold hiện tại):

```
<name>/
├── cmd/                # Entrypoint (main.go)
├── internal/
│   ├── handler/         # HTTP/gRPC handler — chỉ parse input, gọi service, format output
│   ├── service/         # Business logic
│   ├── repository/      # Truy cập database
│   └── model/           # Struct dữ liệu nội bộ
└── Dockerfile
```

- `handler` không chứa business logic — mọi quyết định nghiệp vụ nằm ở `service`.
- `repository` là lớp duy nhất chạm database — `service` không tự viết SQL/query trực tiếp.
- Mỗi service có `Dockerfile` multi-stage riêng (build bằng Go builder image → copy binary vào image `alpine`).

## Vì sao gộp source code dưới `src/`

Toàn bộ `services/`, `frontend/`, `proto/` nằm dưới một thư mục `src/` duy nhất ở root, tách biệt rõ ràng với `docs/`, `api-docs/`, `deployments/` — giữ root repo gọn, dễ phân biệt "thứ được biên dịch/chạy" (`src/`) với "tài liệu và cấu hình vận hành" (mọi thư mục còn lại).
