# FieldSync

A role-based digital platform for standardization and competency-based management of social work field work practicum.

Business requirements (primary source of truth): [`social_work_field_practicum_business_requirements.md`](./social_work_field_practicum_business_requirements.md)
Original platform spec (superseded on workflow specifics — see ARCHITECTURE.md §3a): [`field_sync_requirements.md`](./field_sync_requirements.md)
Architecture and technical decisions: [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md)

## Repository layout

```
mobile/     React Native (Expo) app — Student, Faculty Supervisor, Agency Supervisor
backend/    Go API (Gin) — modular monolith, PostgreSQL via SQLC
docs/       Architecture and planning docs
docker-compose.yml   Local PostgreSQL for development
```

## Prerequisites

- Node.js **>= 20.19.4**
- Go >= 1.22
- Docker Desktop (for local PostgreSQL)
- The mobile app targets **Expo SDK 54** — use a matching Expo Go client
  version, or `npm start` + press `w` to test in a browser instead.

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
# apply migrations in order (PowerShell: use `Get-Content file | docker exec -i ...`
# instead of `<` — PowerShell doesn't support input redirection)
docker exec -i fieldsync-postgres psql -U fieldsync -d fieldsync < migrations/0001_init_users.up.sql
docker exec -i fieldsync-postgres psql -U fieldsync -d fieldsync < migrations/0002_refresh_tokens.up.sql
docker exec -i fieldsync-postgres psql -U fieldsync -d fieldsync < migrations/0003_practicum.up.sql
docker exec -i fieldsync-postgres psql -U fieldsync -d fieldsync < migrations/0004_fieldwork.up.sql
docker exec -i fieldsync-postgres psql -U fieldsync -d fieldsync < migrations/0005_business_model_rework.up.sql
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

Then press `w` for web, or scan the QR code with Expo Go for a device.

If testing on a physical device, `localhost` in `mobile/.env` won't reach
your machine — set `EXPO_PUBLIC_API_URL` to your machine's LAN IP instead
(e.g. `http://192.168.x.x:8090`), and make sure the phone is on the same
Wi-Fi network. A full `npm start` restart is required after changing `.env`
(env vars are baked in at bundler start, not hot-reloaded).

## Authentication

- `POST /auth/register` — self-registration for `student`, `faculty_supervisor`,
  or `agency_supervisor` roles only. Administrator accounts are provisioned
  separately once the Phase 9 Admin dashboard exists — see
  [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md) §8. Requires `institutionId`
  (student/faculty) or `agencyId` (agency supervisor) — every stakeholder is
  tied to a university, directly or via their agency.
- `POST /auth/login` — returns a short-lived JWT access token (15 min) and an
  opaque refresh token (30 days, stored hashed, rotated on every use).
- `POST /auth/refresh` / `POST /auth/logout`
- `GET /users/me` — requires `Authorization: Bearer <accessToken>`
- `GET /public/institutions`, `GET /public/agencies` — unauthenticated, for
  the registration screen's university/agency picker.

The mobile app persists the session in the device's secure storage
(SecureStore) and automatically refreshes an expired access token once
before failing a request.

## Universities & Agencies

Administrator-only management (no Admin UI yet — see [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md) §5b
for how to create an admin user directly in Postgres for local testing).
Agencies are university-scoped (`institutionId` required).

- `POST /institutions`, `GET /institutions`
- `POST /agencies` — body: `name`, `institutionId`; `GET /agencies`
- `GET /agencies/mine` — agencies within the caller's own university.
- `PATCH /agencies/:id`, `DELETE /agencies/:id` — university control over its
  own agency list (body for `PATCH`: `name`).
- `GET /faculty-supervisors/mine` (student) — faculty supervisors within the
  caller's own university.
- `GET /agency-supervisors?agencyId=` — agency supervisors at a given agency.
- `POST /fieldwork-components`, `GET /fieldwork-components` — body: `name`,
  `institutionId`. `PATCH /fieldwork-components/:id`, `DELETE
  /fieldwork-components/:id` — same university-control pattern as agencies.
- `GET /fieldwork-components/mine` (student) — fieldwork components within
  the caller's own university, for the team-request picker.

## Practicum Team Formation

Student-initiated, mutual accept/reject — replaces the earlier
admin-unilateral assignment flow entirely (see
[`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md) §3a #1). The
`Practicum`/`Placement`/`SupervisorAssignment` rows form automatically once
both supervisors accept.

- `POST /team-requests` (student) — body: `agencyId`, `facultySupervisorId`,
  `agencySupervisorId`, `fieldworkComponentId`, `fieldworkDescription`,
  `startDate`. The component must belong to the student's own university.
- `GET /team-requests/me` (student) — the student's own requests
- `GET /team-requests/pending` (faculty/agency supervisor) — requests naming
  the caller, awaiting their decision
- `POST /team-requests/:id/respond` (faculty/agency supervisor) — body:
  `{ "decision": "accepted" | "rejected" }`

Once formed:

- `GET /practicums/me` (student) — active practicum, institution, current
  placement/agency, and assigned supervisors in one response.
- `GET /students` (faculty/agency supervisor) — assigned students with their
  institution and current agency.

## Student Fieldwork

Self-service, student-only, always scoped to the caller's own active
practicum (never a client-supplied student ID). Requires a formed practicum
team, or returns `409 Conflict`.

- `POST /daily-reports` (student, `multipart/form-data`) — fields:
  `reportDate` (`YYYY-MM-DD`), `file` (PDF only, 20 MiB max). The
  handwritten daily fieldwork report (business requirements §10). One per
  (student, date); no edit/resubmit — the doc leaves correction-after-
  rejection explicitly TBD for this record type.
- `GET /daily-reports` (student), `GET /daily-reports/pending` (supervisor,
  same agency-then-faculty visibility rule as attendance).
- `GET /daily-reports/:id/file` — downloads the PDF. Only the owning
  student or an assigned supervisor on that report's practicum may fetch
  it; everyone else gets `403 Forbidden`.
- `POST /daily-reports/:id/agency-review`, `POST /daily-reports/:id/faculty-review`
  — body: `{ "decision": "approved" | "rejected" }`. Same sequential rule as
  attendance review.
- `POST /attendance` — body: `attendanceDate`, `session` (`"morning"` |
  `"evening"`), `hours?` (0–24, optional — hours-to-total calculation is
  explicitly TBD, not computed here). One record per (date, session).
- `GET /attendance` (student) / `GET /attendance/pending` (assigned
  supervisor) — the latter only shows records ready for *that* supervisor:
  agency supervisors see agency-pending records, faculty supervisors only
  see records the agency has already approved.
- `POST /attendance/:id/agency-review`, `POST /attendance/:id/faculty-review`
  — body: `{ "decision": "approved" | "rejected" }`. Faculty review before
  agency approval returns `409 Conflict` (sequential, not independent).
- `POST /consolidated-reports` (student) — body: `summary`. One per
  practicum (`409 Conflict` on a second submission).
- `GET /consolidated-reports/me` (student), `GET /consolidated-reports/pending`
  (supervisor, same agency-then-faculty visibility rule as attendance)
- `POST /consolidated-reports/:id/agency-review`,
  `POST /consolidated-reports/:id/faculty-review` — same sequential rule as
  attendance review.
- `POST /consolidated-reports/:id/resubmit` (student) — body: `summary`.
  Only valid once the report has been rejected by either reviewer; resets
  both review statuses to pending and re-runs the same agency-then-faculty
  approval sequence.

## Feedback

Mandatory weekly feedback from both assigned supervisors, tied to the
practicum record (business requirements §12).

- `POST /feedback` (faculty/agency supervisor) — body: `practicumId`,
  `weekStartDate`, `feedback`. Requires the caller be an assigned supervisor
  on that practicum; one entry per (practicum, supervisor, week).
- `GET /feedback` (student) — all feedback from both supervisors across the
  student's practicum(s).
- `GET /feedback/mine` (supervisor) — feedback the caller has submitted.

## Guidelines & Manuals

One current practicum guidance manual (PDF) per university (business
requirements §17). Versioning/archiving is explicitly TBD, so re-uploading
replaces the previous manual rather than keeping history.

- `POST /manuals` (admin, `multipart/form-data`) — fields: `institutionId`,
  `file` (PDF only, 20 MiB max). Upserts — a second upload for the same
  university replaces the first.
- `GET /manuals` (admin) — list all universities' current manuals.
- `DELETE /manuals/:institutionId` (admin).
- `GET /manuals/mine` (student/faculty/agency supervisor) — the manual for
  the caller's own university (resolved directly for students/faculty,
  via their agency for agency supervisors).
- `GET /manuals/:id/file` — downloads the PDF. Only a member of that
  manual's university (or an administrator) may fetch it; everyone else
  gets `403 Forbidden`.

## Notifications

In-app notifications (business requirements §8, §10 — team request created
and daily report submitted must notify the relevant supervisors; reviewing
a student's own record notifies them too, as a natural extension). No push
notifications yet.

- `GET /notifications` — the caller's own notifications, most recent first.
- `POST /notifications/:id/read` — mark one notification read.
- `POST /notifications/read-all` — mark all of the caller's notifications
  read.

## Development

- `npm run lint` (in `mobile/`) — ESLint via Expo's config
- `go build ./...` / `gofmt -l .` (in `backend/`) — build and format check

## Testing

Backend tests live next to the code they test (`*_test.go`), split into two kinds:

- **Pure unit tests** (`internal/auth/password_test.go`, `jwt_test.go`,
  `middleware_test.go`) — no external dependencies, always run.
- **DB-backed tests** (`service_test.go`, `handler_test.go` in every domain
  package) — require the local Postgres container running
  (`docker compose up -d postgres`). Each test runs inside a transaction
  that's rolled back in cleanup (`internal/testutil.NewTestQueries`), so
  tests never leave data behind or need a separate test database. If
  Postgres isn't reachable, these tests skip (not fail) so `go test ./...`
  still works for someone who hasn't started it yet.

```bash
cd backend
go test ./...          # run everything
go test ./... -v       # verbose, see each test name
```

No mobile test suite yet — revisit once there's mobile-side logic complex
enough to be worth testing beyond typecheck/lint/manual verification.

## Project phases

See [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md) §10 for the original
phase roadmap and §3a for the 2026-08-20 business-model rework that changed
Phase 3 (supervisor assignment → student-initiated team requests) and
Phase 4 (attendance/reports) — both now match the updated business
requirements. See `AGENTS.md` for current status and what's still deferred
(Competency system, Evaluation marks, push notifications — all blocked on
stakeholder-supplied criteria the requirements doc explicitly says not to
invent).
