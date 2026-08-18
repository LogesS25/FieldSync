# FieldSync

A role-based digital platform for standardization and competency-based management of social work field work practicum.

Product requirements: [`social_work_field_practicum_requirements.md`](./social_work_field_practicum_requirements.md)
Architecture and technical decisions: [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md)

## Repository layout

```
mobile/     React Native (Expo) app — Student, Faculty Supervisor, Agency Supervisor
backend/    Go API (Gin) — modular monolith, PostgreSQL via SQLC
docs/       Architecture and planning docs
docker-compose.yml   Local PostgreSQL for development
```

## Prerequisites

- Node.js **>= 20.19.4** (Expo SDK 57 requirement — check with `node --version`)
- Go >= 1.22
- Docker Desktop (for local PostgreSQL)

## Running locally

### 1. Start PostgreSQL

```bash
docker compose up -d postgres
```

This starts Postgres on `localhost:5433` (not the default 5432, to avoid clashing
with other local projects) with database/user/password `fieldsync`.

### 2. Run the backend

```bash
cd backend
cp .env.example .env
# apply the initial migration (Phase 1 only creates the `users` table)
docker exec -i fieldsync-postgres psql -U fieldsync -d fieldsync < migrations/0001_init_users.up.sql
go run ./cmd/api
```

The API listens on `http://localhost:8090`. Verify it's up:

```bash
curl http://localhost:8090/health
# {"status":"ok","database":"connected"}
```

Regenerate SQLC code after changing `sql/queries/*.sql` or `migrations/*.sql`:

```bash
sqlc generate
```

### 3. Run the mobile app

```bash
cd mobile
cp .env.example .env
npm install
npm start
```

Then press `w` for web, or scan the QR code with Expo Go for a device. The
login screen shows live API connectivity status (`API status: ok (db:
connected)`) once the backend and Postgres are running — this is a Phase 1
verification aid, not a real feature.

## Development

- `npm run lint` (in `mobile/`) — ESLint via Expo's config
- `go build ./...` / `gofmt -l .` (in `backend/`) — build and format check

## Project phases

See [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md) §10 for the full roadmap.
This repository is currently at **Phase 1 — Project Foundation**: scaffolding
only, no business features yet.
