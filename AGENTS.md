# AGENTS.md — Hướng dẫn cho AI coding agent

Tài liệu này áp dụng cho mọi AI coding agent (Claude Code, Cursor, Copilot, Codex...) làm việc trên repo TicketFlow. Với các hướng dẫn riêng cho Claude Code, xem [.claude/CLAUDE.md](.claude/CLAUDE.md).

## Trạng thái repo

Phase 1 (MVP) đã được implement trong `src/` theo [docs/07-roadmap/phases-overview.md](docs/07-roadmap/phases-overview.md). Trước khi mở rộng phạm vi, kiểm tra phase hiện tại tại đó và trong domain spec liên quan (`docs/02-domains/<domain>/spec.md`) — không implement vượt phạm vi của phase đang làm.

## Nguồn sự thật (source of truth)

- **Nghiệp vụ & luồng xử lý**: `docs/02-domains/<domain>/spec.md`
- **Schema dữ liệu**: `docs/03-data/` (PostgreSQL DDL, MongoDB document shape, Redis key pattern)
- **Hợp đồng API**: `api-docs/openapi/<service>.yaml` — đây là interface chính thức giữa các service và giữa backend/frontend. Khi implement handler, tuân theo path/schema đã định nghĩa; nếu cần đổi hợp đồng, sửa file OpenAPI trước rồi mới sửa code.
- **Bảo mật & phân quyền**: `docs/04-security/` — đặc biệt lưu ý mục authorization: kiểm tra quyền theo **role** là chưa đủ, còn phải kiểm tra **ownership** (vd: organizer chỉ thao tác được trên sự kiện do chính mình tạo).
- **Quy ước OpenAPI**: [docs/01-architecture/api-conventions.md](docs/01-architecture/api-conventions.md) — style bắt buộc (auth scheme, schema lỗi chung, phân trang, đặt tên path/field) cho mọi file dưới `api-docs/openapi/`.
- **Quy ước backend Go**: [docs/01-architecture/backend-conventions.md](docs/01-architecture/backend-conventions.md) — cách dùng `src/pkg/` (`apperr`, `httpauth`, `authclaims`, `grpcinterceptor`, `pagination`) và quy ước test bắt buộc cho luồng transactional.

## Thêm màn hình / component UI mới

Mọi việc liên quan tới UI/frontend (thêm màn hình mới, thêm component, sửa layout...) phải đọc trước [docs/06-frontend/screens/README.md](docs/06-frontend/screens/README.md) và [docs/06-frontend/components/README.md](docs/06-frontend/components/README.md) — hai file này có quy trình chi tiết (checklist từng bước) phải tuân theo. Nguyên tắc quan trọng nhất, áp dụng cụ thể nguồn sự thật ở mục trên cho phần UI: **rà soát `docs/06-frontend/components/` trước khi mô tả hoặc implement bất kỳ thành phần giao diện nào** — nếu Header/Button/Input/Card... đã có định nghĩa dùng chung, chỉ tham chiếu (tên + variant), không tự vẽ/định nghĩa lại theo cách khác. Token màu/spacing/typography luôn lấy từ [docs/06-frontend/design-system.md](docs/06-frontend/design-system.md); quy tắc hành vi (loading/toast/modal/validate form/pagination) luôn lấy từ [docs/06-frontend/interaction-patterns.md](docs/06-frontend/interaction-patterns.md).

## Quy ước khi sửa đổi

1. Đổi hành vi nghiệp vụ/API/schema → cập nhật spec/OpenAPI tương ứng trong cùng thay đổi.
2. Không tự ý mở rộng phạm vi ngoài domain/phase đang làm; nếu phát hiện thiếu sót trong spec, cập nhật spec rồi mới code theo spec mới — không code trước rồi suy ra spec sau.
3. Mỗi service dưới `src/services/<name>/` sở hữu database riêng (database-per-service) — không truy vấn trực tiếp database của service khác, giao tiếp qua gRPC nội bộ hoặc message queue theo đúng kiến trúc mô tả tại [docs/01-architecture/system-architecture.md](docs/01-architecture/system-architecture.md).
4. Giao tiếp bất đồng bộ giữa các service (đặt vé, thanh toán, thông báo...) tuân theo danh sách event & saga choreography tại [docs/01-architecture/event-driven-design.md](docs/01-architecture/event-driven-design.md) — không tạo thêm event type mới mà không cập nhật tài liệu này.
5. Không viết code "phòng hờ" cho các nice-to-have chưa tới phase — giữ code tối giản đúng scope hiện tại.

## Kiểm tra trước khi coi một thay đổi là hoàn tất

- Với thay đổi API: OpenAPI spec parse hợp lệ và khớp với handler thực tế.
- Với thay đổi schema: migration khớp với `docs/03-data/postgres-schema.md` (hoặc file đã cập nhật).
- Với thay đổi luồng nghiệp vụ có tính transaction (đặt vé, thanh toán): có test cho race-condition/concurrent request, không chỉ happy path.
