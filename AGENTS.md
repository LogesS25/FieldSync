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
- **Navigation is a Drawer (sidebar), not bottom Tabs** — changed 2026-08-21
  after explicit user feedback that bottom tabs with 8 items were unusable
  ("too many tabs... too much for user to find what is theirs"). Both role
  layouts (`(student)/_layout.tsx`, `(supervisor)/_layout.tsx`) use
  `expo-router/drawer`'s `Drawer`, with a shared `useDrawerScreenOptions()`
  hook (`lib/use-drawer-screen-options.ts`) that makes the sidebar a
  slide-out overlay under 768px width and a permanently visible rail at or
  above it — do not add a bottom tab bar again.
- **Reuse the shared form/state components before writing new markup.**
  `components/ui/` now has, beyond the original set:
  `PageHeader` (icon + title + description — the in-body header every real
  screen uses instead of raw `text-2xl` blocks, since the Drawer's top bar
  already carries the route name), `PillSelect` (the recurring "pick one of
  a few" control — role/agency/supervisor/session/student pickers),
  `FormField` (label+error/hint wrapper for non-`TextInput` controls, pairs
  with `PillSelect`/`DateInput`), `LoadingState` and `ErrorState` (every
  query on a real screen should render both, not just the empty-data case —
  before this pass most screens silently showed "nothing here" on a network
  error, which is misleading), and `StatCard` (dashboard summary tiles).
  `components/nav-sidebar.tsx` (`NavSidebar`) is the Drawer's
  `drawerContent` — extend its `ROUTE_ICONS` map when adding a route rather
  than duplicating sidebar markup elsewhere.

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

- **Daily Report file upload (2026-08-21)** — the deferred item from the
  rework above. `field_activities` (free-text daily log) is **replaced
  entirely** by `daily_reports` (business requirements §10's actual "Daily
  Handwritten Report" — a PDF upload), not extended; migration `0007` drops
  `field_activities`/`verification_status` and creates `daily_reports` with
  the same agency-then-faculty sequential review pattern as
  attendance/consolidated reports. Dev-only DB, no real data to migrate
  (2 test rows dropped — see migration 0005/0006 precedent).
  - New `internal/storage` package: local-disk file storage
    (`STORAGE_DIR`, default `./data/uploads`; gitignored). Deliberately
    narrow interface (`Save`/`AbsolutePath`) so swapping to S3 later doesn't
    ripple into `internal/dailyreports` — see production-readiness notes.
    Saved filenames are server-generated (random hex), never derived from
    the student-supplied filename.
  - New `internal/dailyreports` package: `POST /daily-reports` (student,
    multipart form — `reportDate` + `file`, PDF only, 20 MiB cap),
    `GET /daily-reports` (student), `GET /daily-reports/pending`
    (supervisor, same agency-then-faculty visibility as attendance),
    `POST /daily-reports/:id/agency-review`,
    `POST /daily-reports/:id/faculty-review`, and
    `GET /daily-reports/:id/file` (auth-gated download — owning student or
    an assigned supervisor only, `ErrNotYourReport`/`ErrNotAssignedSupervisor`
    otherwise). Unlike the consolidated report, there is **no resubmit** —
    the business doc explicitly leaves daily-report correction/resubmission
    TBD (§10), so building one would have been inventing a business rule.
  - Added `auth.CurrentUserRole` (mirrors the existing `CurrentUserID`
    pattern) since the download-authorization check needed to distinguish
    "owning student" from "assigned supervisor" without a second DB round
    trip for role.
  - `httpserver.NewRouter` signature changed to `(*gin.Engine, error)` since
    router setup now creates the storage directory on boot; `cmd/api`
    updated to handle the error.
  - Full backend test coverage (new `dailyreports` package: sequential
    approval, unassigned-supervisor/wrong-student download rejection,
    duplicate-date rejection, no-active-practicum rejection); verified live
    end-to-end via curl (PDF upload → premature faculty review correctly
    rejected → agency approve → faculty approve → authenticated download
    succeeds → unauthenticated download 401s → pending list empties).
  - Mobile: added `expo-document-picker` (PDF picker), `expo-file-system` +
    `expo-sharing` (native file download/open — the file endpoint requires
    an `Authorization` header, so it can't be a plain link; see
    `lib/open-file.ts`, which branches web-blob-URL vs native-download).
    Added `apiUpload` to `lib/api-client.ts` for multipart bodies (kept
    separate from `apiRequest` — multipart must not be JSON-encoded or have
    a manually-set `Content-Type`). Student and supervisor "Activities" tabs
    (renamed "Daily Reports" in the tab bar; route filenames unchanged)
    replaced entirely — upload form + status list on the student side,
    review queue with a file-view action on the supervisor side. `npx tsc
    --noEmit`, `npm run lint`, and `npx expo export --platform web` all
    clean.

- **Mobile navigation & design rework (2026-08-21)** — full redesign pass
  across the mobile app, not a new feature. User feedback was direct: bottom
  tabs (8 items per role) were unusable, and the visual design needed to be
  "immersive" and "attractive," not just functional. Scope was navigation +
  every screen's loading/error/empty states + spacing, explicitly **not**
  new business logic — see the "Mobile design" standard above for the
  Drawer-navigation and shared-component rules this pass established.
  - Replaced `Tabs` with `expo-router/drawer`'s `Drawer` in both
    `(student)/_layout.tsx` and `(supervisor)/_layout.tsx`; added
    `@react-navigation/drawer` and wrapped the root layout in
    `GestureHandlerRootView` (required for the drawer's gestures — silently
    no-ops without it). New `components/nav-sidebar.tsx` (branded header,
    icon nav list with active-state highlighting, user card + sign-out at
    the bottom) is the shared `drawerContent` for both roles.
  - New shared components (`components/ui/`): `PageHeader`, `PillSelect`,
    `FormField`, `LoadingState`, `ErrorState`, `StatCard` — see the "Mobile
    design" standard above for what each replaces. Existing components
    polished to match (`Card`, `Badge`, `PrimaryButton` — gained a `loading`
    spinner state and an `outline` variant, `EmptyState` — gained an icon,
    `LabeledInput`/`DateInput` — rounded-xl, consistent border/spacing).
    `LogoutButton` gained a `fullWidth` prop for its new home in the sidebar
    footer.
  - Every real (non-placeholder) screen rewritten: both dashboards (the
    supervisor dashboard was a bare placeholder before this — it's now a
    real stat-card summary of assigned students + pending review counts,
    linking into each queue), daily reports, attendance, consolidated
    reports, team/supervision, students list, evaluations — all now handle
    the query-error case explicitly via `ErrorState` with retry, not just
    loading/empty. Placeholder screens (competencies, notifications,
    resources — still Phase 6/7, not built) restyled via a redesigned
    `ScreenPlaceholder` so an unbuilt screen still looks intentional.
    Auth screens (login/register) redesigned with a consistent brand mark
    and the new shared components.
  - Verified via `npx tsc --noEmit`, `npm run lint`, `npx expo export
    --platform web` (all clean, 33 routes), and a live `expo start --web`
    session bundled/served without runtime errors (console/log-checked —
    the Chrome browser tool was unavailable in this environment for a
    visual screenshot check, so this was not eyeballed in-browser; worth a
    manual pass before considering this fully done).

- **Phase 5 — Supervisor Workflows: resolved as complete, nothing left to
  build (2026-08-21).** Reviewed against the current requirements doc with
  the user rather than assumed: review queues + weekly feedback were
  already built in the gap-fill pass; "supervision session" recording
  (from the original spec, FR-15) was explicitly skipped since the
  authoritative requirements doc never re-mentions it (superseded by weekly
  feedback, not silently dropped); Evaluation marks were explicitly skipped
  since criteria/scale/weightage are marked TBD in the requirements doc
  ("do not invent these rules") — same blocked status as Phase 6.
- **Guidelines & Manuals (2026-08-21)** — the well-specified half of Phase
  7 (business requirements §17). Migration `0008`: `manuals` table,
  `UNIQUE(institution_id)` — one current manual per university, replaced
  wholesale on re-upload via `ON CONFLICT DO UPDATE` (upsert), since exact
  versioning/archiving rules are explicitly TBD and this is the simplest
  behavior consistent with what IS specified. New `internal/manuals`
  package, reusing the `internal/storage` package built for daily reports:
  `POST /manuals` (admin, multipart — `institutionId` + PDF file, upsert),
  `GET /manuals` (admin, list all), `DELETE /manuals/:institutionId`
  (admin), `GET /manuals/mine` (any authenticated role — resolves the
  caller's own university in one query: directly via `users.institution_id`
  for students/faculty, indirectly via their agency's `institution_id` for
  agency supervisors), `GET /manuals/:id/file` (auth-gated download — same
  university or administrator only). No admin mobile UI (matches the
  existing agencies/institutions/fieldwork-components precedent — no
  self-service university account exists yet); upload is curl/API-only for
  now. Full backend test coverage (9 tests, including the agency-supervisor
  institution-resolution path); verified live end-to-end via curl (upload →
  replace → student `/manuals/mine` → download 200 → cross-institution
  download correctly 403s).
  Mobile: student and supervisor "Resources" screens (previously
  placeholders) now show the university's manual with a view/download
  action reusing `lib/open-file.ts`, or an empty state if none has been
  uploaded yet. `npx tsc --noEmit`, `npm run lint`, and `npx expo export
  --platform web` all clean.

- **In-app notifications (2026-08-22)** — the other half of Phase 7. Two
  triggers are explicitly mandated by the requirements doc ("the supervisors
  must receive a notification" on team request creation, §8; "the agency
  supervisor and faculty supervisor are notified" on daily report
  submission, §10); the rest (notifying the student when their own
  attendance/daily-report/consolidated-report is reviewed, team-request
  response, team formation, feedback received) is a reasonable UX extension
  of the same mechanism, not new business/scoring rules, so it's included
  too. Migration `0009`: `notifications` table (`recipient_id`, `message`
  — precomputed plain text at insert time, no `kind` enum; add one later if
  the UI needs to visually differentiate types), `created_at DEFAULT
  clock_timestamp()` **not** `now()` — `now()` returns the same value for
  every statement within one transaction, which ties multiple notifications
  created back-to-back (e.g. team formation notifying three people) and
  makes "most recent first" ordering nondeterministic (caught by a real
  test failure, not by inspection).
  - New `internal/notifications` package: `Service.Create` is best-effort
    by design — callers log and swallow a failed insert rather than failing
    the business action that triggered it (a notification bug must never
    block someone from submitting a report). Routes: `GET /notifications`,
    `POST /notifications/:id/read`, `POST /notifications/read-all` — all
    scoped to the caller.
  - Wired into `teamrequests` (create → notify both supervisors; respond →
    notify student; team formed → notify all three), `dailyreports` (create
    → notify both supervisors via new `practicums.ListSupervisorIDs`;
    review → notify student), `attendance` (review → notify student),
    `reports` (review → notify student; resubmit → notify both
    supervisors), and `feedback` (submit → notify student). Every touched
    service's constructor gained a `*notifications.Service` parameter —
    mechanical but real blast radius across 5 packages' tests, all updated.
  - Full backend test coverage (7 new notifications tests; all 5 touched
    packages' existing suites still pass); verified live end-to-end via
    curl through the whole chain (team request → both supervisors notified
    → respond → student notified → team formed → all three notified →
    daily report submitted → supervisors notified → reviewed → student
    notified → mark-read → mark-all-read), using a throwaway backend
    instance on a separate port rather than the user's own running dev
    server.
  - Mobile: new shared `components/notifications-screen.tsx` (both role
    route files render it — no role-specific behavior to justify two
    copies), unread-styled cards, tap-to-mark-read, "mark all read" header
    action. `npx tsc --noEmit`, `npm run lint`, and `npx expo export
    --platform web` all clean.
  - **Not built**: push notifications (needs external Expo push service
    setup — a separate, bigger step) and any missed-deadline/nudge alerting
    (that's a monitoring/compliance concern for Phase 8, not modeled here).

- **Push notifications (2026-08-22)** — device delivery for the in-app
  notifications above, so they reach a phone even when the app is closed.
  Migration `0010`: `push_tokens` table (`user_id`, `token UNIQUE` — a
  device re-registering, e.g. a different user logging in on a shared
  device, reassigns the token via upsert rather than erroring).
  - `internal/notifications` extended, not a new package: `Service.Create`
    now also calls `dispatchPush` after the DB insert. **Concurrency
    lesson, not just a design note**: the first version spawned a goroutine
    that itself queried `*sqlcgen.Queries` for the recipient's tokens. That
    query runs fine against the production connection pool, but every test
    in the whole backend goes through `testutil.NewTestQueries`, which
    wraps the test in a single shared `pgx.Tx` — not safe for concurrent
    access from multiple goroutines. Fixed before it caused flaky tests
    (not after): the token lookup happens synchronously on the caller's
    connection; only the actual outbound HTTP call to Expo's push endpoint
    (no DB access at all) is backgrounded. Rule of thumb for this codebase:
    a goroutine spawned from a `Service` method must never touch
    `s.queries` — hand it plain data instead.
  - New routes: `POST /push-tokens` / `DELETE /push-tokens` (register/
    unregister a device token for the caller). Sending itself
    (`sendExpoPush`) POSTs to Expo's push API (`exp.host/--/api/v2/push/send`)
    with a 10s timeout; failures are logged, never propagated — a
    push-delivery problem must not fail (or even slow down) the request
    that triggered the underlying notification.
  - Full backend test coverage (5 new tests: upsert-reassigns-token,
    unregister, register/unregister handlers, auth-required, and a test
    that `Create` succeeds with zero registered tokens); verified live
    end-to-end via curl including a real (if fake-token) call to Expo's
    actual push endpoint, confirming the server stays healthy and the call
    completes without blocking the triggering request.
  - Mobile: `lib/push-notifications.ts` (`registerForPushNotificationsAsync`
    — requests permission, resolves the Expo push token; returns `null`
    rather than throwing on every failure path: no EAS project configured,
    web, simulator, permission denied), `lib/use-push-registration.ts`
    (registers once per sign-in, from the root layout), unregister wired
    into `LogoutButton`. `npx tsc --noEmit`, `npm run lint`, and
    `npx expo export --platform web` all clean.
  - **This project has no EAS project yet** (confirmed with the user before
    building this), so `registerForPushNotificationsAsync` currently always
    returns `null` in practice — the code is fully wired and correct, but
    won't produce a real, deliverable push token until a human runs `eas
    login` (their own Expo account) and `eas init` from `mobile/`, which
    writes `extra.eas.projectId` into `app.json`. In-app notifications work
    regardless, with or without this step.

### In progress
_(nothing currently)_

### Not started
- Phase 6 — Competency System (framework is an open product decision — do not
  invent scoring rules, keep it configurable — see `docs/ARCHITECTURE.md` §8)
- Evaluation marks (business requirements §14) — same blocking reason as
  Phase 6: criteria/scale/weightage are explicitly TBD, "do not invent
  these rules."
- Phase 8 — Reporting & Monitoring (Basic Requirement Checking, §16, is
  partially blocked too — the "required fieldwork hours" criterion needs
  the hours-calculation formula, which is explicitly TBD per §15; the
  attendance-percentage and required-reports criteria are independently
  checkable since the university supplies its own threshold, not a fixed
  invented value)
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
  Uploaded daily-report PDFs are written to `backend/data/uploads`
  (`STORAGE_DIR`, gitignored) — safe to delete freely in dev.
- Mobile: `npm start` from `mobile/`, press `w` for web or scan the QR with
  Expo Go (must be **SDK 54** — check Expo Go's supported version before
  bumping the Expo SDK). Requires `mobile/.env` with `EXPO_PUBLIC_API_URL` —
  use your LAN IP (not `localhost`) when testing on a physical device.
- Full endpoint list and run instructions: [`README.md`](./README.md).
