---
name: domain-spec
description: Write, update, or audit a docs/02-domains/<domain>/spec.md file against TicketFlow's domain-spec template and cross-doc consistency rules. Use when adding a new domain, changing a domain's business flow/data model, or reviewing an existing domain spec for drift from the template, the OpenAPI contract, the implemented code, or the roadmap.
---

# domain-spec

Viết hoặc audit file `docs/02-domains/<domain>/spec.md`. **Luôn đọc [docs/02-domains/README.md](../../../docs/02-domains/README.md) trước** — đó là nơi chứa bảng chỉ mục, khuôn mẫu chi tiết, và quy trình thêm domain mới. File skill này không lặp lại khuôn mẫu đó, chỉ mô tả cách áp dụng nó qua 2 chế độ.

## Chế độ: viết mới / cập nhật

1. Nếu là domain hoàn toàn mới: xác nhận nó đã có trong bảng service tại `docs/01-architecture/system-architecture.md` — chưa có thì **dừng lại, hỏi người dùng** trước khi viết bất kỳ dòng nào (AGENTS.md: không tự thêm service ngoài kiến trúc đã liệt kê).
2. Xác định phase đang hoạt động qua `docs/07-roadmap/phases-overview.md` và phase-doc tương ứng — không viết nội dung vượt phase đó (AGENTS.md quy ước #2, #5).
3. Đi theo đúng thứ tự "Khuôn mẫu 1 file domain spec" trong `docs/02-domains/README.md`. Với `## Data model`, không mô tả field mà domain khác sở hữu trừ khi đó là carve-out có chủ đích và nêu rõ (mẫu: booking được phép `UPDATE ticket_types.sold_count` dù bảng đó do event-catalog sở hữu).
4. Với section luồng nghiệp vụ: nếu chạm tài nguyên transactional (tồn kho, số dư, trạng thái đơn hàng...), bắt buộc mô tả rõ cơ chế locking + idempotency + compensating transaction khi thất bại/hết hạn — độ chi tiết tham khảo `docs/02-domains/booking/spec.md`. Đây chính là spec mà skill `go-test` sẽ dựa vào để viết race-condition test bắt buộc.
5. Không mô tả một domain ghi trực tiếp vào bảng của domain khác ngoài carve-out đã nêu (database-per-service, AGENTS.md quy ước #3) — giao tiếp chéo domain mô tả qua gRPC/event, nêu trong `## Quan hệ với domain khác`.
6. Nếu luồng phát sinh event type mới, thêm vào bảng tại `docs/01-architecture/event-driven-design.md` trong cùng thay đổi (AGENTS.md quy ước #4) — không tạo event type mà không cập nhật tài liệu đó.
7. Nếu cần endpoint API chưa tồn tại: dừng lại, dùng skill `api-spec` để cập nhật OpenAPI trước, rồi mới viết `## API liên quan` trỏ tới endpoint đó — không mô tả nghiệp vụ dựa trên API chưa được định nghĩa.
8. Thêm/cập nhật dòng tương ứng trong bảng chỉ mục ở `docs/02-domains/README.md`.

## Chế độ: audit

Kiểm tra 1 hoặc nhiều `docs/02-domains/<domain>/spec.md`:

- **Cấu trúc**: đủ heading bắt buộc, đúng thứ tự — header in đậm, `## Phạm vi & trách nhiệm`, `## Quan hệ với domain khác`, `## API liên quan`, `## Phân theo phase` luôn phải có; `## Data model` chỉ được thiếu nếu domain thực sự stateless (như file-storage); section luồng nghiệp vụ có `###` sub-bước nếu nhiều bước.
- **Khớp OpenAPI**: mọi endpoint nhắc tới trong spec phải tồn tại đúng path/method trong `api-docs/openapi/<service>.yaml`; endpoint có trong OpenAPI nhưng domain spec không nhắc tới thì flag để bổ sung nếu đủ quan trọng.
- **Khớp code**: với domain đã có code (`src/services/<service>/`), đối chiếu luồng mô tả (đặc biệt bước locking/transaction) với `internal/repository`/`internal/service` thật — flag khi spec mô tả 1 đằng nhưng code làm 1 nẻo (ví dụ: spec nói khoá theo thứ tự id tăng dần nhưng code không làm vậy).
- **Khớp event**: mọi event type nhắc trong spec phải có trong `docs/01-architecture/event-driven-design.md`.
- **Khớp phase**: bảng `## Phân theo phase` phải nhất quán với `docs/07-roadmap/phase-N-*.md`.
- **Database-per-service**: flag bất kỳ mô tả nào ngụ ý domain A ghi trực tiếp vào bảng domain B mà không phải carve-out đã nêu rõ.
- Báo cáo drift theo dạng: `[domain] <mục cụ thể> — spec nói X, thực tế (code/OpenAPI/roadmap) là Y`.
