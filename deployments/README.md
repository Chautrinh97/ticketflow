# Deployments

**Trạng thái:** chưa có manifest thật — placeholder scaffold.

Cấu hình vận hành hệ thống TicketFlow, tương ứng với lộ trình hạ tầng tại [../docs/07-roadmap/phase-3-infra.md](../docs/07-roadmap/phase-3-infra.md).

```
deployments/
├── docker-compose.yaml   # (sẽ thêm ở Phase 1) chạy toàn bộ hệ thống ở local
├── k8s/
│   ├── base/              # Manifest Kubernetes dùng chung
│   └── overlays/          # Override theo môi trường (dev/prod, kustomize)
└── helm/                  # Helm chart
```

Chi tiết thiết kế hạ tầng: [../docs/05-infra-devops/deployment.md](../docs/05-infra-devops/deployment.md).
