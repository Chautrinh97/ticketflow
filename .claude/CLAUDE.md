# CLAUDE.md

Hướng dẫn dành riêng cho Claude Code khi làm việc trên repo TicketFlow. Đọc [AGENTS.md](../AGENTS.md) ở root trước — file đó chứa quy ước chung áp dụng cho mọi AI coding agent (nguồn sự thật, quy tắc sửa đổi, checklist hoàn tất). Nội dung dưới đây chỉ bổ sung phần đặc thù Claude Code.

## Trước khi implement

- `docs/` là spec, không phải tài liệu tham khảo tuỳ chọn — đọc domain spec liên quan (`docs/02-domains/<domain>/spec.md`) và OpenAPI tương ứng (`api-docs/openapi/<service>.yaml`) trước khi viết handler/model. Nếu công việc thuộc phạm vi UI/frontend, xem thêm mục "Thêm màn hình / component UI mới" trong [AGENTS.md](../AGENTS.md) trước khi mô tả hoặc implement bất kỳ màn hình/component nào.
- Nếu ý định implement khác với spec hiện có (vd: phát hiện spec thiếu case), dừng lại và cập nhật spec trước, hoặc hỏi người dùng nếu không chắc ý định ban đầu.
- Không tự thêm service, bảng dữ liệu, hay endpoint ngoài những gì đã liệt kê trong `docs/01-architecture/system-architecture.md` mà không xác nhận với người dùng.

## Custom subagent của repo này

Xem [.claude/agents/README.md](agents/README.md) — hiện repo dùng các subagent mặc định của Claude Code; subagent chuyên biệt cho từng domain (vd: reviewer cho luồng booking/payment) có thể được thêm khi cần, nay đã có code Phase 1 thật để agent tham chiếu.

## Phạm vi hiện tại

Phase 1 (MVP) đã được implement trong `src/`. Khi implement thêm, chỉ làm trong phạm vi phase đang hoạt động (xem `docs/07-roadmap/`) — không implement vượt trước phase đó.
