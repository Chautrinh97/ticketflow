# CLAUDE.md

Hướng dẫn dành riêng cho Claude Code khi làm việc trên repo TicketFlow. Đọc [AGENTS.md](../AGENTS.md) ở root trước — file đó chứa quy ước chung áp dụng cho mọi AI coding agent (nguồn sự thật, quy tắc sửa đổi, checklist hoàn tất). Nội dung dưới đây chỉ bổ sung phần đặc thù Claude Code.

## Trước khi implement

- `docs/` là spec, không phải tài liệu tham khảo tuỳ chọn — đọc domain spec liên quan (`docs/02-domains/<domain>/spec.md`) và OpenAPI tương ứng (`api-docs/openapi/<service>.yaml`) trước khi viết handler/model.
- Nếu ý định implement khác với spec hiện có (vd: phát hiện spec thiếu case), dừng lại và cập nhật spec trước, hoặc hỏi người dùng nếu không chắc ý định ban đầu.
- Không tự thêm service, bảng dữ liệu, hay endpoint ngoài những gì đã liệt kê trong `docs/01-architecture/system-architecture.md` mà không xác nhận với người dùng.

## Custom subagent của repo này

Xem [.claude/agents/README.md](agents/README.md) — hiện repo dùng các subagent mặc định của Claude Code; subagent chuyên biệt cho từng domain (vd: reviewer cho luồng booking/payment) sẽ được thêm khi bắt đầu Phase 1 implementation, lúc đó có code thật để agent tham chiếu.

## Phạm vi hiện tại

Repo đang ở giai đoạn scaffold tài liệu — **không** implement code trong `src/` trừ khi được yêu cầu rõ ràng và trong phạm vi phase đang làm (xem `docs/07-roadmap/`).
