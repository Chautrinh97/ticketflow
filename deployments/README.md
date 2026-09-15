# Deployments

**Trạng thái:** `docker-compose.yaml` — Phase 1 (MVP). `k8s/`/`helm/` — chưa có, thuộc Phase 3 (xem [../docs/07-roadmap/phase-3-infra.md](../docs/07-roadmap/phase-3-infra.md)).

```
deployments/
├── docker-compose.yaml   # Chạy toàn bộ hệ thống ở local (Phase 1)
├── .env.example          # Copy thành .env (gitignored) để tuỳ chỉnh
├── k8s/
│   ├── base/              # Manifest Kubernetes dùng chung (Phase 3)
│   └── overlays/          # Override theo môi trường (dev/prod, kustomize)
└── helm/                  # Helm chart (Phase 3)
```

## Chạy Phase 1

```bash
cd deployments
cp .env.example .env   # tuỳ chỉnh nếu cần
docker compose up --build
```

Khởi động: `postgres`/`mongo`/`redis` → 4 service `migrate-*` chạy tuần tự (identity → event → booking → payment, mỗi service 1 bảng theo dõi version riêng trên cùng 1 Postgres) → 4 backend service (`identity-service`, `event-service`, `booking-service`, `payment-service`) → `api-gateway` (`:8080`) → `frontend` (`:3000`).

Sau khi mọi container healthy:
```bash
../scripts/seed-dev-users.sh   # seed super_admin/organizer cho dev
../scripts/smoke-test.sh       # kiểm thử end-to-end qua gateway
```

## Port nội bộ giữa các service

4 backend service + `api-gateway` đều lắng nghe **cùng 1 cặp port bên trong container** (`HTTP_PORT=18080`, `GRPC_PORT=18443`) — vì mỗi service là 1 container riêng với network namespace riêng, cái phân biệt chúng với nhau là **tên service** trên `ticketflow-net` (`identity-service`, `event-service`...), không phải số port. Chi tiết xem comment trong `docker-compose.yaml` và `src/services/*/internal/config/config.go`.

4 backend service **không** publish port ra host (chỉ `postgres`/`mongo`/`redis`/`api-gateway`/`frontend` có publish) — mọi truy cập từ bên ngoài đi qua `api-gateway` (`:8080`, map vào port container `18080`). Muốn debug thẳng 1 service (bỏ qua bước kiểm tra session/blacklist của gateway), dùng `docker compose exec`, ví dụ:

```bash
docker compose exec identity-service wget -qO- http://localhost:18080/api/v1/users/me
```

Chi tiết thiết kế hạ tầng production (Phase 3): [../docs/05-infra-devops/deployment.md](../docs/05-infra-devops/deployment.md).
