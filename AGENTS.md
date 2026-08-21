# AGENTS.md — FieldSync

Instructions for any agent (or human) working on this repo. Read this before
making changes. Business requirements (**primary source of truth** — latest
stakeholder discussion, 2026-08-20) live in
[`social_work_field_practicum_business_requirements.md`](./social_work_field_practicum_business_requirements.md).
The original platform spec, [`field_sync_requirements.md`](./field_sync_requirements.md),
is superseded on workflow specifics where the two conflict — see
[`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md) §3a for exactly what changed.
Architecture and technical decisions live in [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md).
This file is about *how* to work in the repo, not *what* the product is.

**Before implementing anything in Phase 3/4/5 territory (supervisor
assignment, attendance, daily reports, consolidated reports, competency
approval, requirement thresholds), read `docs/ARCHITECTURE.md` §3a first** —
that's exactly the area the business requirements doc changed.

---

## Engineering standards (apply from Phase 3 onward)

### Push work into PostgreSQL, don't loop in Go
Prefer a single SQL query (joins, aggregates, `CASE`, window functions,
`FILTER`) over fetching rows and looping/filtering/summing them in
application code. Postgres will always do this faster and with less data
crossing the network. Concretely:
- Listing "students assigned to this supervisor" → one query joining
  `supervisor_assignments` → `students` → `users`, not a fetch-all-then-filter
  in Go.
- Counts, sums, "missing record" detection (Monitoring & Verification module,
  §5.8 of the requirements) → SQL aggregates/`NOT EXISTS` subqueries, not a
  Go loop over all records.
- Only loop in Go over things that are inherently request-shaped (building a
  JSON response from an already-narrow query result) — not to do filtering,
  joining, or aggregation that SQL should own.
- Add indexes for the access patterns each phase actually introduces (FK
  columns used in joins, columns used in `WHERE`/`ORDER BY` on
  hot paths) as part of that phase's migration, not as a later cleanup pass.

### General backend
- Every new domain package follows the existing pattern:
  `handler.go` / `service.go` / migrations + `sql/queries/*.sql` → sqlc.
  `Service` takes `*sqlcgen.Queries` (not a raw pool) — see `internal/auth`
  for the reference shape. This is what makes `internal/testutil.NewTestQueries`
  (transaction-rollback test DB) work everywhere.
- Every new service/handler ships with tests in the same PR/phase, not after:
  pure unit tests for anything with no DB dependency, DB-backed tests via
  `testutil.NewTestQueries` for anything that touches Postgres. Cover the
  edge cases (not just the happy path) — duplicate/conflicting state, wrong
  role, not-found, expired/invalid tokens, empty/malformed input.
- `gofmt -l .`, `go vet ./...`, `go build ./...`, `go test ./...` must all be
  clean before considering a phase done.
- No microservices, no premature abstraction layers. Modular monolith stays
  modular monolith.

### Mobile design
The mobile app should not look like a generic scaffolded CRUD app. Every
screen we build (starting now) should have a considered, immersive visual
design — real attention to layout, spacing, motion, and hierarchy — not just
default `Text`/`View` stacks with Tailwind gray boxes. Placeholder screens
from Phase 1 are expected to look plain; anything built from Phase 3 onward
implementing real functionality is not exempt from this.
- Use NativeWind consistently; build a small shared design vocabulary
  (spacing scale, type scale, a real color palette beyond `slate`/`blue`) as
  reusable components in `components/`, not re-invented per screen.
- Favor clarity and a distinct visual identity over decoration for its own
  sake — this is a professional tool social workers and supervisors use
  daily, not a marketing site. "Immersive" means well-crafted and cohesive,
  not flashy/gimmicky.
- Loading, empty, and error states are part of the design, not an
  afterthought bolted on later.

---

## Status

### Done
- **Phase 1 — Project Foundation**: repo scaffolding (Expo Router + NativeWind
  mobile app, Gin + SQLC Go backend, Docker Compose Postgres), health check,
  local dev docs.
- **Phase 2 — Authentication & User Foundation**: `users` + `refresh_tokens`
  tables; register/login/refresh/logout; JWT access tokens + rotating opaque
  refresh tokens; bcrypt; `RequireAuth`/`RequireRole` middleware; mobile
  login/register screens (RHF + Zod) with SecureStore-persisted session and
  role-based route guards; full backend test suite (40 tests: unit tests for
  password/JWT/middleware, DB-backed tests for service/handler layers via
  transaction-rollback test isolation).
- **Phase 3 — Practicum & Placement**: `institutions`, `agencies`,
  `practicums`, `placements`, `supervisor_assignments` tables;
  administrator-gated create/list endpoints (no Admin UI yet — see §5b of
  `docs/ARCHITECTURE.md` for creating a test admin user directly in
  Postgres); `GET /practicums/me` and `GET /students` as single-round-trip
  queries using `LATERAL` joins + `json_agg` instead of Go-side
  looping/joining (see Engineering standards above); real mobile screens
  (student dashboard, supervisor students list) replacing the Phase 1
  placeholders, plus a small shared design system (`components/ui/`: Card,
  Badge, Avatar, EmptyState, ScreenContainer; `brand`/`accent` color tokens
  in `tailwind.config.js`); backend test suite covers service + handler
  layers including role-mismatch and duplicate-assignment edge cases.
- **Phase 4 — Student Fieldwork**: `field_activities`, `attendance_records`,
  `weekly_reports` tables (`verification_status`/`report_status` enums ready
  for Phase 5 to use, nothing sets them beyond the default yet). Student-only
  self-service endpoints, always scoped to the caller's own active practicum
  via `auth.CurrentUserID` — never a client-supplied student ID. Two scope
  decisions made explicitly (documented in `docs/ARCHITECTURE.md` §5c, not
  silently assumed): no edit/delete on activities/attendance, and weekly
  reports submit in one action rather than draft-then-submit. Field-hours
  total computed via `SUM(...)` in Postgres. Real mobile screens for
  activities/attendance/reports, plus a couple of shared components
  (`LabeledInput`, `PrimaryButton`) added to the design system. Two bugs
  found via manual testing and fixed: login/register weren't navigating away
  from the login screen after success (store updated but router never
  moved), and an `administrator` account would have infinite-looped between
  `/` and the `(supervisor)` group's role guard. **Follow-up after user
  testing**: the first pass shipped plain text inputs (`YYYY-MM-DD` typed by
  hand) for every date field — a real design-quality miss against the
  "immersive, not boring" bar set for Phase 3 onward. Fixed same-day, not
  deferred: added `components/ui/date-input.tsx` (native picker via
  `@react-native-community/datetimepicker`, iOS wrapped in a bottom-sheet
  with Cancel/Done since iOS has no dismissible date-only dialog) +
  `date-input.web.tsx` (native browser `<input type="date">`, platform file
  resolved automatically by Metro). Lesson: when a screen has a date field,
  reach for `DateInput` from the start — a raw text field for a date is not
  an acceptable placeholder going forward, not even for a first pass.

- **Business Model Rework (2026-08-20/21)** — done, see `docs/ARCHITECTURE.md`
  §3a for the full change log this implements. Migration `0005`:
  - `agencies` gained `institution_id` (NOT NULL) — agencies are now
    university-scoped, not global. `GET /agencies/mine`,
    `GET /public/agencies` added.
  - New `practicum_team_requests` table + `internal/teamrequests` package:
    student picks agency/faculty supervisor/agency supervisor, each responds
    independently (`POST /team-requests`, `/team-requests/me`,
    `/team-requests/pending`, `/team-requests/:id/respond`); the practicum
    team (`Practicum`+`Placement`+both `SupervisorAssignment`s) forms
    automatically the moment both accept. The old admin-unilateral
    `POST /practicums` / `/placements` / `/supervisor-assignments` HTTP
    routes were removed (the underlying `practicums.Service` methods remain,
    now called internally by `teamrequests`).
  - `attendance_records` reworked: added `session` (`morning`/`evening`,
    unique per date+session, replacing one-record-per-day), and
    `agency_status`/`faculty_status` review columns replacing the unused
    single `verification_status`. Faculty review is rejected with `409` if
    attempted before agency approval (`ErrAgencyReviewFirst`) — sequential,
    not independent. `hours_logged` became optional; the old
    `/attendance/summary` total-hours endpoint was **removed entirely**,
    not just hidden — hours-to-total calculation is explicitly TBD in the
    business doc and computing one would have been inventing a business
    rule.
  - `weekly_reports` **dropped**, replaced by `consolidated_reports`: one row
    per practicum (`UNIQUE(practicum_id)`), same agency-then-faculty
    sequential review pattern as attendance.
  - Registration now requires `institutionId` (student/faculty_supervisor)
    or `agencyId` (agency_supervisor) — previously nothing set
    `users.institution_id`/`agency_id` at all, which would have made the
    team-request flow unreachable for any freshly-registered user. Added
    `GET /public/institutions`, `GET /public/agencies` (unauthenticated, for
    the registration picker) and `GET /faculty-supervisors/mine`,
    `GET /agency-supervisors?agencyId=` (for the team-request picker).
  - Full backend test coverage for all of the above (new tests in
    `agencies`, `attendance`, `reports`, `teamrequests`, `users`; existing
    `auth`/`practicums` tests updated for the new required registration
    fields). Verified live end-to-end via curl: full team formation, faculty
    review before agency approval correctly rejected, sequential
    attendance and consolidated-report approval, duplicate-report rejection.
  - Mobile: registration screen gained university/agency pickers (was a hard
    regression otherwise — registration would always fail without them);
    new student "Team" screen (`(student)/supervision.tsx`) for creating and
    tracking team requests; reworked attendance screen (session picker, no
    more fake total-hours card) and reports screen (single consolidated
    report, not a list); new supervisor screens for team-request
    accept/reject, attendance review, and consolidated-report review
    (`(supervisor)/supervision.tsx`, `attendance.tsx`, `evaluations.tsx`).
    `PrimaryButton` gained a `variant` prop (`brand`/`danger`/`neutral`) so
    Accept and Reject don't look identical.
  - **Explicitly deferred, not started**: `FieldActivity` → `DailyReport`
    (business doc's "daily handwritten report" is a file upload, not free
    text) — needs a real file storage backend first (local disk for dev;
    was Phase 7 scope), pulled forward as a dependency but not yet built.
    `field_activities`/`GET /field-activities` still work exactly as
    before (Phase 4 semantics), they just don't yet match the business
    doc's file-upload requirement.

- **Business Model Rework gap-fill (2026-08-21)** — implements the four
  concrete gaps left after the rework above. Migration `0006`:
  - New `fieldwork_components` table (`institution_id`, `name`,
    `UNIQUE(institution_id, name)`) + `internal/fieldworkcomponents` package:
    full admin-gated CRUD (`POST/GET/PATCH/:id/DELETE /fieldwork-components`)
    plus `GET /fieldwork-components/mine` (student, scoped to own
    institution). `practicum_team_requests` gained a required
    `fieldwork_component_id` FK — `teamrequests.CreateRequest` validates the
    component exists and belongs to the student's own institution before
    insert, rather than trusting the client-supplied ID.
  - New `weekly_feedback` table (`practicum_id`, `supervisor_id`,
    `week_start_date`, `feedback`, `UNIQUE(practicum_id, supervisor_id,
    week_start_date)`) + `internal/feedback` package: either assigned
    supervisor submits (`POST /feedback`, validated against
    `SupervisorAssignmentExists` first), student reads all of theirs
    (`GET /feedback`), supervisor reads their own submissions
    (`GET /feedback/mine`).
  - Consolidated-report resubmission: `Service.Resubmit` requires the report
    belong to the caller and be currently rejected by at least one reviewer
    (`ErrNotYourReport`/`ErrNotRejected`), then resets both review statuses
    to pending and re-runs the same agency-then-faculty approval sequence.
    Route: `POST /consolidated-reports/:id/resubmit`.
  - University CRUD over its own reference lists: added `PATCH`/`DELETE
    /agencies/:id` (admin-gated, scoped to the caller's institution).
    Deliberately **not** built in this pass (flagged, not silently skipped):
    attendance-requirements/checkbox-conditions/manuals CRUD (those entities
    don't exist at all yet) and faculty-supervisor-list management (would
    mean deactivating user accounts — a separate, larger decision).
  - Full backend test coverage (new `fieldworkcomponents`/`feedback`
    packages, extended `teamrequests`/`agencies`/`reports` tests); verified
    live end-to-end via curl (component rename visible to student, team
    request with component forms correctly, feedback POST visible to
    student, reject→resubmit→re-approve cycle).
  - Mobile: student "Team" screen gained a fieldwork-component picker
    (same pill-list pattern as the other pickers on that screen); supervisor
    "Supervision" screen gained a Weekly Feedback form (student picker +
    `DateInput` week-start + text, using the shared components); student
    "Team" screen gained a Weekly Feedback received list, most-recent-first;
    student "Reports" screen gained a resubmit form that only appears when
    the current report has been rejected by either reviewer. `npx tsc
    --noEmit`, `npm run lint`, and `npx expo export --platform web` all
    clean.

### In progress
_(nothing currently — gap-fill complete, Phase 5 not yet started)_

### Not started
- **Daily Report file upload** (deferred from the rework above — see Done).
  Needs: local-disk `storage.Storage` implementation, multipart upload
  endpoint, mobile image picker (`expo-image-picker` or
  `expo-document-picker`) + upload screen. Do this before or alongside
  Phase 5, since Phase 5's "review daily records" workflow logically
  includes reviewing these once they exist.
- Phase 5 — Supervisor Workflows (review, verification, supervision records,
  feedback, evaluations)
- Phase 6 — Competency System (framework is an open product decision — do not
  invent scoring rules, keep it configurable — see `docs/ARCHITECTURE.md` §8)
- Phase 7 — Resources & Notifications (S3 file storage, push notifications)
- Phase 8 — Reporting & Monitoring
- Phase 9 — Administrator web dashboard (Next.js, same Go API)

---

## Production readiness — required before this app goes live

This app will have real users (students, faculty, agency supervisors)
handling real practicum data. The items below are known, accepted gaps for
the current MVP/dev stage — they are not optional polish, they must be
addressed before a public launch. Nobody should assume "it works in my dev
environment" means "safe to expose to real users."

### Backend security
- **`JWT_SECRET`**: `.env.example` ships a placeholder
  (`dev-only-insecure-secret-change-me`). Every real environment needs its
  own strong random secret from a real secret manager (AWS Secrets Manager,
  etc.) — never a checked-in file, never reused across environments.
- **CORS**: currently `AllowAllOrigins: true` (see
  `internal/httpserver/router.go`) so the Expo web dev target can call the
  API from any port. Must become an explicit origin allowlist before any
  public deployment.
- **Rate limiting**: `/auth/login`, `/auth/register`, `/auth/refresh` have no
  rate limiting — currently open to brute-force/credential-stuffing and
  registration spam. Needs per-IP and/or per-account limiting.
- **Email verification**: self-registration (`POST /auth/register`) accepts
  any email without confirming ownership. Needs a verification flow before
  going live, or at minimum before treating an account as fully trusted.
- **Account lockout / anomalous-login detection**: not implemented.
- **Password reset ("forgot password")**: does not exist yet. Users who
  forget their password have no recovery path.
- ~~**`GIN_MODE`**~~ — **done**: `httpserver.NewRouter` calls
  `gin.SetMode(gin.ReleaseMode)` automatically when `APP_ENV=production`.
- **Trusted proxies**: Gin logs a warning about trusting all proxies by
  default — must configure the real proxy chain (or disable) once deployed
  behind a load balancer/reverse proxy.
- **TLS**: the Go process itself speaks plain HTTP. Needs a TLS-terminating
  proxy/load balancer in front, and all mobile traffic must go over HTTPS in
  production (`EXPO_PUBLIC_API_URL` currently points at plain `http://` for
  local dev).

### Backend operations
- **Migrations**: currently applied by hand via `psql`/`docker exec`. Before
  production, adopt a versioned migration runner (e.g. `golang-migrate`)
  with a tracked schema-version table so deploys are repeatable and
  reversible, and multiple environments can't drift out of sync.
- **Logging/monitoring**: only Gin's default stdout logger exists. Needs
  structured logging and an error-monitoring service (e.g. Sentry) before
  real incidents need debugging.
- **Database backups**: no backup/restore strategy defined yet for the
  production Postgres instance.
- **Secrets management**: `.env` files are fine for local dev only —
  production secrets (DB credentials, JWT secret, future S3 keys) need a
  real secret store, never a file shipped in a deploy artifact.

### Mobile / product
- **Mobile route guards are UX only, not security**: `RequireRole` in
  `components/require-role.tsx` prevents the app from *rendering* the wrong
  screens — it is not, and must never be treated as, the actual
  authorization boundary. The backend's `RequireAuth`/`RequireRole`
  middleware is the real boundary; the mobile guard exists purely so a
  logged-in user with the wrong role sees a sensible redirect instead of a
  broken screen.
- **Data sensitivity / compliance**: this app will hold social work
  practicum records — potentially sensitive personal data depending on
  jurisdiction. Privacy policy, data retention policy, and any applicable
  compliance requirements (FERPA-adjacent for student records, etc.) are
  explicitly called out as unresolved in the requirements doc §13/§14 and
  must get real legal/product review before launch — this is not something
  to assume or invent in code.
- **Crash reporting / analytics**: not set up.
- **App store requirements**: privacy policy and data-collection disclosures
  for App Store/Play Store review are not addressed.

## Local dev environment (for reference)
- Postgres: Docker container `fieldsync-postgres`, port **5433** (not 5432 —
  already taken by another local project). DBeaver connection:
  `localhost:5433`, db/user/password all `fieldsync`.
- Backend: `go run ./cmd/api` from `backend/`, listens on port **8090** (not
  8080 — same reason). Requires `backend/.env` (copy from `.env.example`).
- Mobile: `npm start` from `mobile/`, press `w` for web or scan the QR with
  Expo Go (must be **SDK 54** — check Expo Go's supported version before
  bumping the Expo SDK). Requires `mobile/.env` with `EXPO_PUBLIC_API_URL` —
  use your LAN IP (not `localhost`) when testing on a physical device.
- Full endpoint list and run instructions: [`README.md`](./README.md).
