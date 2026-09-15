# Domains — chỉ mục & khuôn mẫu

Mỗi domain có 1 file spec riêng (`<domain>/spec.md`), viết theo khuôn mẫu ở mục "Khuôn mẫu 1 file domain spec" bên dưới. File domain spec là **nguồn sự thật cho nghiệp vụ & luồng xử lý** (xem [../../AGENTS.md](../../AGENTS.md#nguồn-sự-thật)) — đọc trước khi viết handler/model, và trước khi viết `api-docs/openapi/<service>.yaml` cho service tương ứng.

## Bảng chỉ mục

| Domain | Service sở hữu | Database | Phase | File |
|---|---|---|---|---|
| Booking | Booking Service | PostgreSQL (`orders`, `order_items`, `tickets`) | 1 (MVP) | [booking/spec.md](booking/spec.md) |
| Event Catalog | Event Service | PostgreSQL (`events`, `ticket_types`) + MongoDB (`event_catalog`) | 1 (CRUD cơ bản) → 2 (full-text/fuzzy search) | [event-catalog/spec.md](event-catalog/spec.md) |
| Identity | Identity Service | PostgreSQL (`users`) | 1 (auth cơ bản) → 2 (duyệt organizer + RBAC đầy đủ) | [identity/spec.md](identity/spec.md) |
| Payment | Payment Service | PostgreSQL (`payments`) | 1 (mock) → 2 (cổng thanh toán thật) | [payment/spec.md](payment/spec.md) |
| Notification | Notification Service | PostgreSQL (`notifications`) | 2 (Must-have) | [notification/spec.md](notification/spec.md) |
| File Storage | File Service | không có (stateless) | 2 (Must-have) | [file-storage/spec.md](file-storage/spec.md) |
| Analytics | Booking/Event Service (Phase 2, query trực tiếp) → Analytics Service riêng (Phase 4) | PostgreSQL (Phase 2), ClickHouse (Phase 4) | 2 (cơ bản) → 4 (Analytics Service độc lập) | [analytics/spec.md](analytics/spec.md) |
| Search | Search Service | Elasticsearch | 4 (Nice-to-have); Phase 1-2 nằm trong [event-catalog](event-catalog/spec.md) | [search/spec.md](search/spec.md) |
| AI Integration | chưa xác định tên cụ thể (endpoint bổ sung trong Event Service + service mỏng riêng) | PostgreSQL + `pgvector` (hoặc dense vector của Elasticsearch) | 4 (Nice-to-have) | [ai-integration/spec.md](ai-integration/spec.md) |

## Khuôn mẫu 1 file domain spec

Mọi file `<domain>/spec.md` viết theo đúng thứ tự sau:

1. **Header**: dòng in đậm `**Service sở hữu:** ... · **Database:** ... · **Phase:** ...` ngay dưới tiêu đề `# Domain: <Tên>`.
2. **`## Phạm vi & trách nhiệm`**: domain này chịu trách nhiệm gì, không chịu trách nhiệm gì (ranh giới với domain khác).
3. **`## Data model`** *(bỏ qua nếu domain không sở hữu bảng/collection riêng, vd file-storage)*: liệt kê bảng/collection sở hữu, link `docs/03-data/`; nêu rõ nếu có cột "sổ cái" domain khác sở hữu nhưng domain này được phép cập nhật trong transaction của chính nó (mẫu: `event-catalog` sở hữu `ticket_types` nhưng `booking` là bên duy nhất `UPDATE sold_count`).
4. **Section luồng nghiệp vụ** (tên tuỳ domain — vd "Luồng đặt vé (booking flow)", "Luồng nghiệp vụ", "Đồng bộ dữ liệu"/"Truy vấn", hoặc chia theo `## Giai đoạn Phase N` nếu domain rẽ nhánh rõ theo phase): dùng `###` cho sub-luồng nhiều bước. Nếu luồng đụng tài nguyên transactional (tồn kho, số dư...), **bắt buộc** nêu rõ cơ chế locking, idempotency, và compensating transaction khi thất bại/hết hạn — độ chi tiết tham khảo [booking/spec.md](booking/spec.md).
5. **`## Quan hệ với domain khác`**: liệt kê theo domain liên quan, nêu đích danh field/event nào đi qua ranh giới (không mô tả chung chung "có liên quan tới X").
6. **`## API liên quan`**: link `../../../api-docs/openapi/<service>.yaml`; nếu domain dùng chung service với domain khác (vd analytics dùng chung Booking/Event Service ở Phase 2), nêu rõ endpoint nào thuộc phạm vi domain này.
7. **`## Phân theo phase`**: bảng `Tính năng | Phase`, khớp với `docs/07-roadmap/phase-N-*.md`.

## Quy trình thêm domain mới

1. Xác nhận domain đã có trong bảng service tại [../01-architecture/system-architecture.md](../01-architecture/system-architecture.md) — **chưa có thì dừng lại, xác nhận phạm vi với người dùng trước** (theo [../../AGENTS.md](../../AGENTS.md), không tự thêm service ngoài kiến trúc đã liệt kê).
2. Xác định phase qua [../07-roadmap/phases-overview.md](../07-roadmap/phases-overview.md) — không viết spec vượt phase đang làm.
3. Viết `## Data model` tham chiếu `docs/03-data/` — giữ field thuộc sở hữu đúng service, không mô tả service khác trực tiếp ghi vào bảng của domain này (ngoại lệ có chủ đích thì nêu rõ, theo mẫu ở mục 3 phía trên).
4. Viết section luồng nghiệp vụ; nếu chạm tài nguyên transactional, mô tả locking + idempotency + compensating transaction (xem [booking/spec.md](booking/spec.md) làm mẫu).
5. Điền `## Quan hệ với domain khác` và `## API liên quan`. Nếu cần endpoint mới, cập nhật `api-docs/openapi/<service>.yaml` theo [../01-architecture/api-conventions.md](../01-architecture/api-conventions.md) trong cùng thay đổi — không viết spec dựa trên API chưa tồn tại.
6. Nếu luồng phát sinh event type mới, thêm vào bảng tại [../01-architecture/event-driven-design.md](../01-architecture/event-driven-design.md) trong cùng thay đổi.
7. Thêm 1 dòng vào "Bảng chỉ mục" ở đầu file này.
