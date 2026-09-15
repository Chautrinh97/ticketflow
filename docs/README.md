# TicketFlow — Tài liệu thiết kế

Mục lục điều hướng toàn bộ tài liệu spec của TicketFlow. Đây là **nguồn sự thật** cho thiết kế hệ thống — code triển khai phải khớp với tài liệu tại đây; khi nghiệp vụ thay đổi, cập nhật tài liệu trước hoặc cùng lúc với code.

## 00 — Tổng quan

- [project-overview.md](00-overview/project-overview.md) — Bài toán, mục tiêu dự án
- [roles-permissions.md](00-overview/roles-permissions.md) — Vai trò người dùng & ma trận phân quyền
- [tech-stack.md](00-overview/tech-stack.md) — Công nghệ sử dụng theo từng lớp hệ thống

## 01 — Kiến trúc

- [system-architecture.md](01-architecture/system-architecture.md) — Danh sách service, trách nhiệm, giao tiếp
- [repo-structure.md](01-architecture/repo-structure.md) — Cấu trúc thư mục repo
- [event-driven-design.md](01-architecture/event-driven-design.md) — Message event, saga choreography

## 02 — Domain spec

Mỗi domain map 1-1 với một service (trừ vài domain nice-to-have chưa tách service riêng):

- [identity/spec.md](02-domains/identity/spec.md)
- [event-catalog/spec.md](02-domains/event-catalog/spec.md)
- [booking/spec.md](02-domains/booking/spec.md)
- [payment/spec.md](02-domains/payment/spec.md)
- [notification/spec.md](02-domains/notification/spec.md)
- [file-storage/spec.md](02-domains/file-storage/spec.md)
- [search/spec.md](02-domains/search/spec.md) *(nice-to-have)*
- [analytics/spec.md](02-domains/analytics/spec.md)
- [ai-integration/spec.md](02-domains/ai-integration/spec.md) *(nice-to-have)*

## 03 — Thiết kế dữ liệu

- [postgres-schema.md](03-data/postgres-schema.md)
- [mongodb-schema.md](03-data/mongodb-schema.md)
- [redis-keys.md](03-data/redis-keys.md)

## 04 — Bảo mật

- [authentication.md](04-security/authentication.md)
- [authorization.md](04-security/authorization.md)
- [rate-limiting.md](04-security/rate-limiting.md)

## 05 — Hạ tầng & DevOps

- [deployment.md](05-infra-devops/deployment.md)
- [background-jobs.md](05-infra-devops/background-jobs.md)
- [observability.md](05-infra-devops/observability.md) *(nice-to-have)*

## 06 — Frontend

- [README.md](06-frontend/README.md) — tổng quan kỹ thuật, render mode, auth phía client
- [design-system.md](06-frontend/design-system.md) — token màu/typography/spacing (theo thang Tailwind)
- [interaction-patterns.md](06-frontend/interaction-patterns.md) — quy tắc hành vi dùng chung (loading, toast, modal, validate form...)
- [components/](06-frontend/components/) — component dùng chung giữa các màn hình
- [screens/](06-frontend/screens/) — đặc tả chi tiết từng màn hình

## 07 — Lộ trình triển khai

- [phases-overview.md](07-roadmap/phases-overview.md)
- [phase-1-mvp.md](07-roadmap/phase-1-mvp.md)
- [phase-2-must-have.md](07-roadmap/phase-2-must-have.md)
- [phase-3-infra.md](07-roadmap/phase-3-infra.md)
- [phase-4-nice-to-have.md](07-roadmap/phase-4-nice-to-have.md)

## 08 — Hướng mở rộng tương lai

- [future-directions.md](08-future/future-directions.md)

## Tài liệu liên quan (ngoài `docs/`)

- Hợp đồng API: [../api-docs/](../api-docs/)
- Cấu trúc source code: [../src/](../src/)
