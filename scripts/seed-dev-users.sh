#!/usr/bin/env bash
# Seeds a super_admin and an organizer account directly into Postgres, per
# docs/07-roadmap/phase-1-mvp.md: "có thể seed sẵn vài tài khoản
# organizer/super_admin để phát triển" — Phase 1 has no self-service
# organizer-approval flow yet (that's Phase 2).
#
# Usage: ./scripts/seed-dev-users.sh
# Requires: psql, and postgres already migrated (docker compose up, with
# the migrate-* one-shot services having completed).
set -euo pipefail

DATABASE_URL="${DATABASE_URL:-postgres://ticketflow:ticketflow@localhost:5432/ticketflow?sslmode=disable}"

psql "$DATABASE_URL" -v ON_ERROR_STOP=1 <<'SQL'
INSERT INTO users (id, email, full_name, role, status)
VALUES (gen_random_uuid(), 'admin@ticketflow.dev', 'Super Admin', 'super_admin', 'active')
ON CONFLICT (email) DO NOTHING;

INSERT INTO users (id, email, full_name, role, status)
VALUES (gen_random_uuid(), 'organizer@ticketflow.dev', 'Demo Organizer', 'organizer', 'active')
ON CONFLICT (email) DO NOTHING;
SQL

cat <<EOF
Seeded dev accounts:
  super_admin: admin@ticketflow.dev
  organizer:   organizer@ticketflow.dev

Log in via the frontend's mock login with either email (any full name) —
AUTH_FIREBASE_MODE=mock links the account by email on first login (see
src/services/identity-service/internal/service/auth_service.go).
EOF
