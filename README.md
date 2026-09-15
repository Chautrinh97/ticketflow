# TicketFlow

Nền tảng đặt vé sự kiện trực tuyến (concert, workshop, thể thao...) theo kiến trúc microservice — backend Go, frontend Next.js, hạ tầng Kubernetes.

TicketFlow cho phép các đơn vị tổ chức (organizer) đăng sự kiện có số lượng vé giới hạn, và người dùng tìm kiếm, đặt & thanh toán vé trực tuyến. Ba yêu cầu kỹ thuật cốt lõi mà hệ thống phải đáp ứng:

- **Không bán trùng vé** khi nhiều người đặt cùng lúc (flash-sale) — đảm bảo bằng transaction ACID + locking ở tầng dữ liệu.
- **Schema linh hoạt theo loại sự kiện** (concert cần nghệ sĩ, workshop cần giảng viên, thể thao cần đội thi đấu) bên cạnh dữ liệu giao dịch chặt chẽ.
- **Tìm kiếm chịu lỗi chính tả** (full-text + fuzzy search) theo tên sự kiện, địa điểm.

## Trạng thái hiện tại

**Phase 1 (MVP) đã implement** — xem [docs/07-roadmap/phase-1-mvp.md](docs/07-roadmap/phase-1-mvp.md): 4 backend service Go (identity, event, booking, payment) + API Gateway, chạy được end-to-end bằng `docker-compose` (đăng nhập → xem sự kiện → đặt vé nhiều loại vé/1 đơn với transaction ACID chống bán trùng → thanh toán mock → vé được phát hành), cùng frontend Next.js cho 12 màn hình cơ bản. Chi tiết cách chạy: [deployments/README.md](deployments/README.md). Các phase sau (RBAC đầy đủ, search, hạ tầng Kubernetes, event-driven...) xem [docs/07-roadmap/phases-overview.md](docs/07-roadmap/phases-overview.md).

## Bắt đầu từ đâu

| Muốn biết... | Xem tại |
|---|---|
| Bức tranh tổng thể dự án, mục tiêu | [docs/00-overview/project-overview.md](docs/00-overview/project-overview.md) |
| Vai trò người dùng & phân quyền | [docs/00-overview/roles-permissions.md](docs/00-overview/roles-permissions.md) |
| Công nghệ sử dụng | [docs/00-overview/tech-stack.md](docs/00-overview/tech-stack.md) |
| Kiến trúc hệ thống, danh sách service | [docs/01-architecture/system-architecture.md](docs/01-architecture/system-architecture.md) |
| Spec chi tiết từng domain nghiệp vụ | [docs/02-domains/](docs/02-domains/) |
| Thiết kế dữ liệu (Postgres/Mongo/Redis) | [docs/03-data/](docs/03-data/) |
| Bảo mật, auth, rate limit | [docs/04-security/](docs/04-security/) |
| Hạ tầng, DevOps, cron job | [docs/05-infra-devops/](docs/05-infra-devops/) |
| Frontend (tổng quan, render mode) | [docs/06-frontend/README.md](docs/06-frontend/README.md) |
| Spec chi tiết từng màn hình UI + component dùng chung | [docs/06-frontend/screens/](docs/06-frontend/screens/), [docs/06-frontend/components/](docs/06-frontend/components/) |
| Lộ trình triển khai theo phase | [docs/07-roadmap/](docs/07-roadmap/) |
| Hợp đồng API (OpenAPI) | [api-docs/](api-docs/) |
| Cấu trúc thư mục source code | [src/](src/) |

## Cấu trúc repo

```
.
├── docs/           # Tài liệu spec: overview, kiến trúc, domain, data, security, infra, frontend, roadmap
├── api-docs/       # Hợp đồng API (OpenAPI 3.0) cho từng service
├── src/            # Source code (services Go, frontend Next.js, proto dùng chung) — Phase 1 (MVP) đã implement
├── deployments/    # Docker Compose, Kubernetes manifests, Helm chart
├── scripts/        # Script tiện ích cho dev/CI
├── .claude/        # Cấu hình Claude Code cho repo này
└── AGENTS.md       # Quy ước chung cho AI coding agent làm việc trên repo
```

## Quy tắc làm việc

- **`docs/` là nguồn sự thật (source of truth)** cho thiết kế — trước khi implement bất kỳ tính năng nào, đọc domain spec tương ứng trong `docs/02-domains/`.
- Thay đổi hành vi nghiệp vụ/API/schema → cập nhật spec liên quan trong cùng lần thay đổi, không để tài liệu lệch khỏi thực tế.
- Xem [AGENTS.md](AGENTS.md) để biết quy ước chi tiết khi dùng AI coding agent trên repo này.
