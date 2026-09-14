# Custom subagent cho TicketFlow

Thư mục này sẽ chứa các custom subagent (`.md` với frontmatter `name`/`description`/`tools`) dành riêng cho repo TicketFlow, bổ sung cho các subagent mặc định của Claude Code.

## Trạng thái hiện tại

Chưa có subagent nào được định nghĩa — repo mới ở giai đoạn scaffold tài liệu, chưa có code để agent chuyên biệt review/thao tác trên đó.

## Dự kiến bổ sung từ Phase 1 (khi có code)

| Subagent dự kiến | Mục đích |
|---|---|
| `booking-flow-reviewer` | Review code liên quan transaction đặt vé — soi các trường hợp race-condition, thiếu `SELECT ... FOR UPDATE`, thiếu rollback/compensating transaction |
| `api-contract-checker` | Đối chiếu handler thực tế với `api-docs/openapi/*.yaml`, phát hiện lệch hợp đồng |
| `domain-spec-writer` | Hỗ trợ soạn thảo/cập nhật spec trong `docs/02-domains/` khi nghiệp vụ thay đổi |

Khi thêm subagent mới, cập nhật bảng trên và giữ mỗi subagent tập trung vào một trách nhiệm rõ ràng thay vì tạo một agent "làm tất cả".
