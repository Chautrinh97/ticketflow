# DevOps & Deployment

## Docker

Mỗi service một `Dockerfile` multi-stage: build bằng Go builder image → copy binary sang image `alpine` nhỏ gọn. Frontend Next.js dùng multi-stage tương tự (build → `node:alpine` runtime hoặc export static nếu phù hợp).

## Kubernetes

- Mỗi service: `Deployment` (nhiều replica) + `Service` (ClusterIP) + `ConfigMap`/`Secret` riêng.
- `Ingress` (Nginx hoặc Traefik) làm reverse proxy kiêm load balancer, route theo path/subdomain tới từng service.
- `HorizontalPodAutoscaler` scale theo CPU/RAM.
- Liveness/readiness probe (`/healthz`) bắt buộc cho mọi service — readiness probe phải phản ánh đúng khả năng phục vụ (vd: kết nối được DB), không chỉ trả `200` cố định.
- Dùng **k9s** để quan sát/debug pod trong quá trình phát triển và demo.

Manifest tổ chức theo kustomize: `deployments/k8s/base/` chứa cấu hình chung, `deployments/k8s/overlays/{dev,prod}/` override theo môi trường (số replica, resource limit, domain...). Song song đó, `deployments/helm/` cung cấp cách triển khai bằng Helm chart cho ai muốn dùng Helm thay vì kustomize.

## SSL

`cert-manager` tự động xin & gia hạn chứng chỉ Let's Encrypt cho Ingress — không tự quản lý chứng chỉ thủ công.

## Cloudflare

DNS đặt ở chế độ proxy (orange cloud) để có CDN + WAF + DDoS protection + rate-limit rule ở biên, trước khi request chạm tới cluster. Xem thêm [../04-security/rate-limiting.md](../04-security/rate-limiting.md) về vai trò bổ sung (không thay thế) cho rate-limit ở API Gateway.

## CI/CD (GitHub Actions)

Pipeline chuẩn: `lint → unit test → build Docker image → push registry → kubectl apply / helm upgrade`.

- **lint**: chạy `golangci-lint` cho mỗi service Go thay đổi, `eslint`/`tsc` cho frontend.
- **unit test**: chạy test của service/package thay đổi; với Booking Service, bắt buộc có test race-condition (nhiều goroutine đặt vé đồng thời) trước khi cho phép merge.
- **build & push**: build image gắn tag theo commit SHA, push registry (GitHub Container Registry hoặc tương đương).
- **deploy**: `kubectl apply -k deployments/k8s/overlays/<env>` hoặc `helm upgrade` — chỉ tự động deploy môi trường `dev`; deploy `prod` cần approval thủ công (environment protection rule của GitHub Actions).

## Local development

`deployments/docker-compose.yaml` khởi chạy toàn bộ hệ thống (các service, Postgres, MongoDB, Redis, message broker khi tới Phase 4) ở máy local, dùng cho phát triển hằng ngày mà không cần Kubernetes.
