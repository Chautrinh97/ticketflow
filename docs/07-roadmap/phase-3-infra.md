# Phase 3 — Hạ tầng production

**Ước lượng:** 1–2 tuần. **Mục tiêu:** hệ thống (đã đủ tính năng must-have từ Phase 2) chạy được trên hạ tầng giống production thật, thể hiện năng lực vận hành chứ không chỉ code.

## Phạm vi

| Hạng mục | Việc cần làm | Tài liệu liên quan |
|---|---|---|
| Kubernetes | Viết `Deployment`/`Service`/`ConfigMap`/`Secret` cho từng service, deploy bằng Helm chart | [../05-infra-devops/deployment.md](../05-infra-devops/deployment.md) |
| Ingress + SSL | Nginx/Traefik Ingress, `cert-manager` cấp chứng chỉ Let's Encrypt tự động | [../05-infra-devops/deployment.md](../05-infra-devops/deployment.md) |
| Cloudflare | DNS proxy mode, WAF, rate-limit rule ở biên | [../04-security/rate-limiting.md](../04-security/rate-limiting.md) |
| Autoscaling & health check | `HorizontalPodAutoscaler` theo CPU/RAM, liveness/readiness probe (`/healthz`) cho mọi service | [../05-infra-devops/deployment.md](../05-infra-devops/deployment.md) |
| CI/CD | Pipeline GitHub Actions đầy đủ: lint → test → build → push image → deploy (tự động cho `dev`, cần approval cho `prod`) | [../05-infra-devops/deployment.md](../05-infra-devops/deployment.md) |
| Demo vận hành | Dùng k9s quan sát pod, giả lập tải cao (flash-sale) và chỉ ra hệ thống scale/đứng vững | [../05-infra-devops/deployment.md](../05-infra-devops/deployment.md) |

## Không thay đổi logic nghiệp vụ ở phase này

Phase 3 là hạ tầng thuần tuý — không thêm/sửa tính năng nghiệp vụ. Nếu trong lúc deploy phát hiện thiếu sót nghiệp vụ (vd: thiếu health check phản ánh đúng trạng thái DB), sửa đúng phạm vi đó, không mở rộng thêm domain mới.

## Ngoài phạm vi Phase 3

Message queue/saga, observability (tracing/metrics/log tập trung), Elasticsearch, AI — thuộc [Phase 4](phase-4-nice-to-have.md).
