# Custom subagent cho TicketFlow

Thư mục này sẽ chứa các custom subagent (`.md` với frontmatter `name`/`description`/`tools`) dành riêng cho repo TicketFlow, bổ sung cho các subagent mặc định của Claude Code.

## Trạng thái hiện tại

Chưa có subagent riêng nào được định nghĩa cho repo này — hiện dùng các subagent mặc định của Claude Code. Phase 1 (MVP) đã có code thật (5 Go service + frontend Next.js), nhưng phần "review/thao tác chuyên biệt trên code" được đưa vào **skill** dưới [.claude/skills/](../skills/) thay vì subagent — xem mục dưới.

## Dự kiến ban đầu (đã gộp vào skill — xem mục "Cập nhật" bên dưới)

| Subagent dự kiến | Mục đích |
|---|---|
| `booking-flow-reviewer` | Review code liên quan transaction đặt vé — soi các trường hợp race-condition, thiếu `SELECT ... FOR UPDATE`, thiếu rollback/compensating transaction |
| `api-contract-checker` | Đối chiếu handler thực tế với `api-docs/openapi/*.yaml`, phát hiện lệch hợp đồng |
| `domain-spec-writer` | Hỗ trợ soạn thảo/cập nhật spec trong `docs/02-domains/` khi nghiệp vụ thay đổi |

Bảng trên giữ lại chỉ để tham khảo lịch sử quyết định — không còn là kế hoạch đang thực hiện.

## Cập nhật: gộp vào skill thay vì subagent

Quyết định cuối: đóng gói cả phần "viết mới" lẫn "audit" dưới dạng skill (`.claude/skills/<name>/SKILL.md`), không tạo subagent riêng, để nhất quán và đơn giản hơn khi maintain. 3 subagent dự kiến ở trên tương ứng với:

- `booking-flow-reviewer` → [go-audit](../skills/go-audit/SKILL.md) (soi race-condition/locking/rollback) kết hợp [go-test](../skills/go-test/SKILL.md) (đảm bảo có test race-condition).
- `api-contract-checker` → [api-spec](../skills/api-spec/SKILL.md) ở chế độ audit (đối chiếu YAML↔handler), chạy cùng [go-audit](../skills/go-audit/SKILL.md) để có coverage đủ 2 chiều.
- `domain-spec-writer` → [domain-spec](../skills/domain-spec/SKILL.md) ở chế độ viết mới/cập nhật.

Skill chỉ chạy khi được gọi (theo mô tả hoặc `/tên-skill`), không tự động chạy liên tục như một subagent review-per-change từng hình dung ban đầu. Nếu sau này cần audit tự động mỗi lần đổi code (pre-commit hook, wrapper command...), cân nhắc bổ sung ở một việc riêng — chưa nằm trong phạm vi hiện tại.

Khi thêm subagent hoặc skill mới, giữ mỗi cái tập trung vào một trách nhiệm rõ ràng thay vì tạo một agent/skill "làm tất cả".
