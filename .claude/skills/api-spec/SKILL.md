---
name: api-spec
description: Write, update, or audit an api-docs/openapi/<service>.yaml file against TicketFlow's OpenAPI house style, and check it for drift against the corresponding domain spec and the real Go handler code. Use when adding/changing an API endpoint contract, or reviewing an existing OpenAPI file for style violations or handler mismatch.
---

# api-spec

Viết hoặc audit `api-docs/openapi/<service>.yaml`. **Luôn đọc [docs/01-architecture/api-conventions.md](../../../docs/01-architecture/api-conventions.md) trước** — đó là nơi chứa toàn bộ quy ước bắt buộc và quy trình. File skill này mô tả cách áp dụng quy ước đó qua 2 chế độ.

## Chế độ: viết mới / cập nhật

1. Xác định domain sở hữu qua `docs/02-domains/README.md`, tìm đúng file `.yaml` tương ứng trong bảng chỉ mục ở `api-conventions.md`.
2. Sửa OpenAPI **trước khi code handler** (AGENTS.md quy ước #1) — nếu task tới từ yêu cầu implement code, dừng lại và cập nhật YAML trước.
3. Nếu là endpoint hoàn toàn mới, ngoài phạm vi đã ngụ ý trong `docs/01-architecture/system-architecture.md`, dừng lại và xác nhận với người dùng trước khi thêm.
4. Áp dụng đúng mọi quy ước trong `api-conventions.md`: `servers: [{url: /api/v1}]`; `security: bearerAuth` global + `security: []` override cho endpoint public; response lỗi luôn `$ref: '#/components/schemas/Error'`; endpoint list luôn dùng `PageParam`/`PageSizeParam` + response `{items,total}`; path resource-plural + verb sub-resource; field snake_case; **không** thêm `operationId`; mô tả tiếng Việt trỏ về domain spec.
5. Cập nhật mục `## API liên quan` trong domain spec tương ứng nếu cần (dùng skill `domain-spec` nếu thay đổi đáng kể).
6. Đọc lại toàn bộ path/schema vừa sửa để tự kiểm tra cấu trúc YAML hợp lệ (`$ref` trỏ đúng, không thiếu response, không trùng operation) — repo chưa có tool validate tự động.

## Chế độ: audit

Kiểm tra 1 hoặc nhiều file `.yaml` (có thể đối chiếu với code thật khi service đã implement):

- **Style**: mọi quy ước liệt kê ở `api-conventions.md` — flag riêng từng vi phạm (vd: thiếu `security: []` cho endpoint đáng lẽ public, response lỗi không dùng `$ref: Error`, dùng `page`/`page_size` khác tên, path dùng verb, field camelCase, hoặc — hiếm nhưng phải bắt — có `operationId` xuất hiện).
- **Khớp handler thật** (khi `src/services/<service>/internal/handler/http/` đã có code): với mỗi path+method trong YAML, tìm route tương ứng trong `router.go`/handler — flag path/method có trong YAML nhưng không có route, và ngược lại route có trong code nhưng YAML không khai báo (đây là chức năng thay thế cho subagent `api-contract-checker` từng dự kiến, xem `.claude/agents/README.md`).
- **Khớp field**: đối chiếu field trong `requestBody`/response schema với struct Go tương ứng trong `internal/model` hoặc struct request/response của handler — flag field thiếu/thừa/sai kiểu.
- **Khớp security thật**: đối chiếu `security: []`/`security: [bearerAuth]` khai báo với middleware chain thật trên route đó (`httpauth.RequireAuth`/`RequireRole`/`RequireOwnership`) — flag endpoint khai báo public nhưng code có `RequireAuth`, hoặc ngược lại.
- Chạy cùng skill `go-audit` khi audit 1 service để có coverage đầy đủ cả 2 chiều (YAML→code và code→YAML).
- Báo cáo drift theo dạng: `[service] <path> <method> — YAML nói X, code thực tế là Y`.
