CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organizer_id UUID NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    category VARCHAR(50) NOT NULL,
    venue_name VARCHAR(255),
    address TEXT,
    city VARCHAR(100),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    banner_url TEXT,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    search_vector tsvector,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_events_search ON events USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_events_trgm_title ON events USING GIN (title gin_trgm_ops);

CREATE TABLE IF NOT EXISTS ticket_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id),
    name VARCHAR(100) NOT NULL,
    price NUMERIC(12,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'VND',
    quota INT NOT NULL,
    sold_count INT NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 0
);
