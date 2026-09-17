# PostgreSQL — schema chính

Dữ liệu giao dịch chặt chẽ (user, sự kiện phần cấu trúc cứng, đơn hàng, vé, thanh toán) nằm ở PostgreSQL, đảm bảo ACID. Mỗi service sở hữu một phần schema riêng theo nguyên tắc database-per-service — các bảng dưới đây được nhóm theo service sở hữu để rõ ranh giới, dù có thể cùng nằm trong một cluster Postgres ở giai đoạn MVP (tách instance vật lý khi cần scale độc lập).

## Identity Service — `users`

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    firebase_uid VARCHAR(128) UNIQUE,
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(255),
    avatar_url TEXT,
    role VARCHAR(20) NOT NULL DEFAULT 'user', -- super_admin | organizer | user
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active | pending | banned
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
```

`status='pending'` dùng cho tài khoản vừa đăng ký trở thành `organizer`, chờ `super_admin` duyệt — xem [../02-domains/identity/spec.md](../02-domains/identity/spec.md).

## Event Service — `events`, `ticket_types`

```sql
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organizer_id UUID NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    category VARCHAR(50) NOT NULL, -- concert | workshop | sport
    venue_name VARCHAR(255),
    address TEXT,
    city VARCHAR(100),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    banner_url TEXT,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft', -- draft | published | cancelled
    search_vector tsvector,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_events_search ON events USING GIN (search_vector);
CREATE INDEX idx_events_trgm_title ON events USING GIN (title gin_trgm_ops); -- fuzzy search

CREATE TABLE ticket_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id),
    name VARCHAR(100) NOT NULL, -- VIP | Standard...
    price NUMERIC(12,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'VND',
    quota INT NOT NULL,
    sold_count INT NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 0 -- dùng cho optimistic locking nếu không dùng SELECT ... FOR UPDATE
);
```

`search_vector` phục vụ full-text search, `gin_trgm_ops` (extension `pg_trgm`) phục vụ fuzzy search — chi tiết tại [../02-domains/event-catalog/spec.md](../02-domains/event-catalog/spec.md) và [redis-keys.md](redis-keys.md) (cache liên quan).

## Booking Service — `orders`, `order_items`, `tickets`

```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending | paid | cancelled | expired
    total_amount NUMERIC(12,2) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL, -- huỷ tự động nếu quá hạn (cron job)
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id),
    ticket_type_id UUID NOT NULL REFERENCES ticket_types(id),
    quantity INT NOT NULL,
    unit_price NUMERIC(12,2) NOT NULL
);

CREATE TABLE tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_item_id UUID NOT NULL REFERENCES order_items(id),
    ticket_code VARCHAR(64) UNIQUE NOT NULL, -- dùng để sinh QR
    status VARCHAR(20) NOT NULL DEFAULT 'valid', -- valid | used | cancelled
    issued_at TIMESTAMPTZ DEFAULT now()
);
```

`tickets` chỉ được tạo **sau khi thanh toán thành công** — trước đó chỉ có `orders`/`order_items` ở trạng thái `pending`. Chi tiết transaction/locking khi tạo đơn xem [../02-domains/booking/spec.md](../02-domains/booking/spec.md).

### `organizer_revenue_daily_snapshots` (Phase 2)

```sql
CREATE TABLE organizer_revenue_daily_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organizer_id UUID NOT NULL REFERENCES users(id),
    snapshot_date DATE NOT NULL,
    total_revenue NUMERIC(14,2) NOT NULL DEFAULT 0,
    total_orders INT NOT NULL DEFAULT 0,
    tickets_sold INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (organizer_id, snapshot_date)
);
```

Ghi bởi cron job "Báo cáo doanh thu ngày" ([../05-infra-devops/background-jobs.md](../05-infra-devops/background-jobs.md)) — mỗi organizer có tối đa một dòng snapshot/ngày (`UNIQUE (organizer_id, snapshot_date)`), job phải dùng `INSERT ... ON CONFLICT DO UPDATE` để chạy lại an toàn (idempotent) nếu job bị retry cùng ngày. Mục đích: tránh quét lại toàn bộ `orders`/`order_items` mỗi lần xem lịch sử doanh thu theo ngày — xem [../02-domains/analytics/spec.md](../02-domains/analytics/spec.md).

## Payment Service — `payments`

```sql
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id),
    provider VARCHAR(50), -- stripe | vnpay | momo...
    provider_txn_id VARCHAR(128),
    amount NUMERIC(12,2) NOT NULL,
    status VARCHAR(20) NOT NULL, -- initiated | success | failed
    raw_payload JSONB,
    created_at TIMESTAMPTZ DEFAULT now()
);
```

`provider_txn_id` dùng để xử lý webhook idempotent — xem [../02-domains/payment/spec.md](../02-domains/payment/spec.md).

## Notification Service — `notifications`

```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255),
    body TEXT,
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now()
);
```

## Cross-cutting — `audit_logs`

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT now()
);
```

Ghi lại các hành động nhạy cảm (khoá/mở khoá tài khoản, duyệt organizer, đổi role, huỷ vé bởi admin...) — chỉ `super_admin` xem được, xem [../00-overview/roles-permissions.md](../00-overview/roles-permissions.md).

## Ghi chú về ranh giới database-per-service

Các bảng trên có khoá ngoại tham chiếu chéo (vd: `events.organizer_id` → `users.id`) để đơn giản hoá schema ở giai đoạn thiết kế/MVP khi chạy chung một cluster Postgres. Khi tách instance vật lý theo service (Phase 3+), các tham chiếu chéo này chuyển thành việc lưu `id` dạng UUID không có ràng buộc `FOREIGN KEY` thật (chỉ còn là quy ước), và việc xác thực tồn tại của resource ở service khác thực hiện qua gRPC thay vì constraint database.
