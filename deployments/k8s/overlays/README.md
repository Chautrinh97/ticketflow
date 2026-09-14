# K8s overlays

**Trạng thái:** chưa có overlay — placeholder scaffold.

Sẽ chứa hai overlay (kustomize) override manifest ở [../base/](../base/) theo môi trường:

- `dev/` — số replica thấp, resource limit nhỏ, domain `*.dev.ticketflow...`
- `prod/` — số replica/resource theo nhu cầu thật, domain chính thức, bật `HorizontalPodAutoscaler`

Chi tiết: [../../../docs/05-infra-devops/deployment.md](../../../docs/05-infra-devops/deployment.md).
