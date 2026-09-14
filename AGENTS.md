# AGENTS.md — Hướng dẫn cho AI coding agent

Tài liệu này áp dụng cho mọi AI coding agent (Claude Code, Cursor, Copilot, Codex...) làm việc trên repo TicketFlow. Với các hướng dẫn riêng cho Claude Code, xem [.claude/CLAUDE.md](.claude/CLAUDE.md).

## Trạng thái repo

Repo đang ở giai đoạn **pre-implementation**: `docs/` và `api-docs/` chứa đầy đủ spec, `src/` mới chỉ có khung thư mục (README mô tả ý định, chưa có code). Khi bắt đầu code, hãy đối chiếu với phase hiện tại trong [docs/07-roadmap/phases-overview.md](docs/07-roadmap/phases-overview.md) trước — không implement vượt phạm vi của phase đang làm.

## Nguồn sự thật (source of truth)

- **Nghiệp vụ & luồng xử lý**: `docs/02-domains/<domain>/spec.md`
- **Schema dữ liệu**: `docs/03-data/` (PostgreSQL DDL, MongoDB document shape, Redis key pattern)
- **Hợp đồng API**: `api-docs/openapi/<service>.yaml` — đây là interface chính thức giữa các service và giữa backend/frontend. Khi implement handler, tuân theo path/schema đã định nghĩa; nếu cần đổi hợp đồng, sửa file OpenAPI trước rồi mới sửa code.
- **Bảo mật & phân quyền**: `docs/04-security/` — đặc biệt lưu ý mục authorization: kiểm tra quyền theo **role** là chưa đủ, còn phải kiểm tra **ownership** (vd: organizer chỉ thao tác được trên sự kiện do chính mình tạo).

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
