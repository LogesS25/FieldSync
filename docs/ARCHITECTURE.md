# FieldSync — Architecture

Source of truth for product requirements: [`social_work_field_practicum_requirements.md`](../social_work_field_practicum_requirements.md).
This document does not restate or reinterpret those requirements — it records the
technical decisions made to implement them, and flags what is still open.

## 1. Product phasing

- **Phase 1 (current)**: mobile app foundation only. Roles: Student, Faculty
  Supervisor, Agency Supervisor. Administrator is explicitly out of scope for
  the mobile app.
- **Later**: Next.js Administrator web dashboard, consuming the same Go API.

The backend is built API-first and role-agnostic at the data layer so the
Administrator dashboard can be added later without restructuring the core
system.

## 2. High-level architecture

```
React Native (Expo) ──REST──> Go API (Gin) ──> PostgreSQL
                                            └──> S3 (via storage.Storage interface)
```

- **Modular monolith.** No microservices. Domain logic is split into Go
  packages under `internal/`, each with its own handler/service/repository,
  but all deployed as one binary.
- **Separation of concerns.** HTTP handlers translate transport ↔ domain;
  services hold business rules; repositories (SQLC-generated) hold SQL.
- **API-first.** Mobile and (later) web are both just REST clients.

## 3. Domain model (derived from requirements, not assumed)

Entities identified from the requirements' stakeholders, modules, and workflow
sections:

| Entity | Purpose |
|---|---|
| `User` | Base identity: email, password hash, role, timestamps |
| `Institution` | Educational institution a student/faculty belongs to |
| `Agency` | Field placement organization |
| `Practicum` | A student's enrollment period, links Student ↔ Institution |
| `Placement` | Links a Practicum to an Agency |
| `SupervisorAssignment` | Links a Faculty/Agency Supervisor to a Student's Practicum — this is the authorization boundary for "assigned students" |
| _(implementation note)_ | `institution_id`/`agency_id` live directly on `users` (nullable FKs) rather than separate `StudentProfile`/`FacultySupervisorProfile`/`AgencySupervisorProfile` tables as originally sketched — simpler schema, fewer joins, and there are no profile-specific fields yet to justify separate tables. Revisit if/when role-specific fields appear (e.g. agency supervisor qualification credentials, §4.3). |
| `FieldActivity` | Daily logged field work; has verification status + verifier |
| `AttendanceRecord` | Attendance entries with verification status |
| `WeeklyReport` | Aggregated weekly submission; submission state |
| `SupervisionSession` | Recorded supervision meeting |
| `Evaluation` | Structured score + criteria reference, authored by a supervisor |
| `Feedback` | Freeform feedback linked to an activity/report/session |
| `CompetencyFramework` / `CompetencyCriterion` | **Configurable**, not hardcoded — see Open Decisions |
| `CompetencyScore` | A student's progress against criteria |
| `Resource` | File metadata (manuals, templates, guidelines); actual bytes in S3 |
| `Notification` | Directed alert to a specific stakeholder |
| `Grievance` | Raised by an agency supervisor |
| `VerificationIssue` | System-detected gap (computed, not directly written by users) |

Authorization is enforced by walking these relationships (e.g., a Faculty
Supervisor may only act on a Student reachable via their
`SupervisorAssignment` rows), not by a flat permissions table.

## 4. Backend

- **Language/framework**: Go + Gin.
- **Database access**: SQLC (type-safe generated Go from raw SQL) + pgx/v5
  driver. No ORM — keeps generated SQL auditable and avoids hidden N+1s.
- **Package layout**: `internal/<domain>/{handler,service,repository,model}.go`
  per bounded context (auth, users, students, practicums, placements,
  activities, attendance, reports, supervision, evaluations, competencies,
  resources, notifications, grievances, verification), plus shared
  `httpserver`, `config`, `db`, `storage` packages.
- **Auth**: JWT access tokens + refresh tokens (Phase 2). Authentication
  (who are you) and authorization (what can you do) are handled as separate
  middleware/service layers.
- **File storage**: `internal/storage.Storage` interface (`Upload` /
  `Download` / `Delete`). Production implementation targets S3; local dev can
  use a filesystem-backed implementation so contributors don't need AWS
  credentials to run the app. Only the interface exists in Phase 1.

### Why Go + Gin over alternatives?
Chosen per the project brief — the backend has meaningful business logic
(RBAC, workflow state, verification, competency tracking) that benefits from
Go's simplicity and strong typing, and Gin is a lightweight, well-understood
router. Alternative considered: Node/Express (team already knows JS from the
mobile side) — rejected per explicit instruction to use Go, and because a
single shared language across mobile/backend was not a stated goal.

### Why SQLC over an ORM (e.g., GORM, Ent)?
SQL stays explicit and reviewable; generated code is a thin, typed wrapper
with no runtime query-building magic. Trade-off: SQLC requires migrations to
exist before generating code, so schema and queries evolve together
deliberately rather than being inferred from structs.

## 5. Database

- **PostgreSQL**, matching the highly relational domain.
- **Migrations**: plain numbered `.up.sql` / `.down.sql` files in
  `backend/migrations/`. Phase 1 added the `users` table (with a `user_role`
  enum); Phase 2 added `refresh_tokens`. The full domain schema (institutions,
  practicums, activities, competencies, etc.) is designed in Phase 3+
  alongside the corresponding business logic, not upfront.

## 5a. Authentication (Phase 2)

- **Password hashing**: bcrypt.
- **Access tokens**: JWT (HS256), 15-minute TTL, carries `sub` (user ID) and
  `role`. Stateless — authorization middleware validates the signature and
  expiry only, no DB lookup per request.
- **Refresh tokens**: opaque random tokens, stored **hashed** (SHA-256) in
  `refresh_tokens`, 30-day TTL. Rotated on every use — refreshing revokes the
  presented token and issues a new pair, so a leaked refresh token stops
  working the moment the legitimate client refreshes.
- **Self-registration** (`POST /auth/register`) is allowed for `student`,
  `faculty_supervisor`, and `agency_supervisor` only. This was an explicit
  decision (not specified by the requirements, which imply Administrator-led
  provisioning — see §8) made to unblock testing before the Phase 9 Admin
  dashboard exists. `administrator` is hard-excluded from the registerable
  role set in code (`auth.RegisterableRoles`), not just UI-hidden — revisit
  this endpoint's exposure/rate-limiting before any real deployment.
- **Mobile session storage**: SecureStore (device Keychain/Keystore) via a
  Zustand `persist` middleware adapter, not plain AsyncStorage.

## 5b. Practicum & Placement (Phase 3)

- **Who can create institutions/agencies/practicums/placements/supervisor
  assignments?** Administrator-only (`POST /institutions`, `/agencies`,
  `/practicums`, `/placements`, `/supervisor-assignments`, all gated by
  `RequireRole("administrator")`). This matches the requirements (§4.4 —
  Administrator manages institutions, assigns supervisors) rather than
  inventing a self-service flow. Since there's no Admin UI until Phase 9,
  an administrator account has to be inserted directly into Postgres for
  now (self-registration as `administrator` is blocked — see §5a); the API
  endpoints are real and tested, just not yet reachable from the mobile app.
- **Query design leans on Postgres, not Go loops**: `GetPracticumSummaryForStudent`
  resolves the student's current placement via a `LATERAL` join and
  aggregates assigned supervisors with `json_agg` — one round trip, no
  application-side joining. `ListStudentsForSupervisor` similarly resolves
  each student's current agency via `LATERAL` rather than an N+1 fetch. See
  `backend/sql/queries/practicums.sql`. This is now the standard pattern for
  any query that would otherwise require looping over a result set to
  join/aggregate — see `AGENTS.md`.
- `GET /practicums/me` (student) and `GET /students` (faculty/agency
  supervisor) are the read endpoints the mobile app actually uses.

## 5c. Student Fieldwork (Phase 4)

Two explicit scope decisions, made because the requirements don't specify
either and inventing an unconfirmed workflow was judged worse than a
narrower but honest MVP:

- **No edit/delete for field activities or attendance.** Requirements §14
  Q13 ("Can students edit records after submission?") is an open question.
  Phase 4 only implements create + list. `verification_status` exists on
  both tables (defaulting to `pending`) so Phase 5's supervisor verification
  can build on this schema without another migration, but nothing sets it
  to `verified`/`rejected` yet.
- **Weekly reports are submitted in a single action, not drafted.** FR-07
  only requires "submit weekly reports" — a draft-then-submit state machine
  isn't specified anywhere. `report_status` still has `submitted`/`reviewed`
  values so Phase 5 can add the review action without a migration.

Ownership is enforced at the handler level: every write derives the
student's identity from the JWT (`auth.CurrentUserID`), never from a
client-supplied ID — a student cannot create a record for another student
by any input manipulation. Writes require an active practicum
(`practicums.Service.GetActivePracticumID`); attempting to log
activity/attendance/reports without one returns `409 Conflict`.

`GET /attendance/summary` computes total field hours with `SUM(...)` in
Postgres (`internal/attendance` `GetTotalHoursForStudent` query) rather than
fetching all records and summing in Go — same pattern as §5b.

## 6. Mobile

- **Expo (managed workflow) + React Native + TypeScript.**
- **Navigation**: Expo Router, route groups by auth state and role:
  `(auth)`, `(student)`, `(supervisor)`. One app, not three — the backend
  returns the user's role on login and the app renders the matching
  navigation tree. Faculty vs. Agency Supervisor share the `(supervisor)`
  group since their capability lists in the requirements are structurally
  similar (assigned students, verification, evaluation, feedback); the
  specific actions available within a screen are gated by role at the
  component level.
- **Styling**: NativeWind (Tailwind for RN).
- **Server state**: TanStack Query for all API-backed data.
- **Client state**: Zustand, scoped to auth/session and transient UI state
  only — no server data duplicated into Zustand stores.
- **Forms**: React Hook Form + Zod schemas, mirrored (not shared, since
  validation must be authoritative server-side too) by Go server-side
  validation.

## 7. Notifications

Domain model only in Phase 1 scope planning (`Notification` entity in the
schema design); no delivery mechanism (push, email) is built until Phase 7,
per the roadmap. Avoids building notification infrastructure before there is
a concrete trigger (verification alerts, missing-record alerts) to wire it to.

## 8. Open product decisions (NOT assumed, per requirements §13/§14)

These require stakeholder input before the corresponding features can be
finalized. The architecture keeps each of these as **configuration/data**,
not hardcoded logic, specifically so they can be resolved later without a
rearchitecture:

1. Exact competency framework, categories, and scoring formula.
2. Minimum required field hours.
3. Report submission cadence (weekly is named, but confirm strictly weekly).
4. Required supervision frequency.
5. Agency supervisor qualification/verification criteria.
6. Whether institutions may customize requirements vs. what must stay global.
7. Rejection workflow: what happens when a supervisor rejects a submission.
8. Whether students can edit records after submission, and until when.
9. Full approval/rejection state machine for reports/activities.
10. Which reports are mandatory vs. optional.
11. Who may view a given student's competency records (visibility rules
    beyond "assigned supervisor").
12. Grievance/complaint handling process and resolution workflow.
13. Notification channels required (push only, or also email/SMS) and
    per-channel delivery rules.

## 9. Reasonable technical decisions made now (not product requirements)

- JWT access + refresh token auth (industry-standard fit for a mobile-first
  API; no stated alternative in requirements).
- UUID primary keys (avoids leaking sequential IDs across institutions/
  agencies in a multi-tenant-ish domain).
- Postgres enum for `user_role` in Phase 1; may move to a lookup table if
  roles need to become dynamically configurable later.
- Local dev Postgres via Docker Compose, single `postgres` service for now.

## 10. Development phases

See the roadmap below; each phase builds on the previous and does not start
until the prior phase's foundation is in place.

1. **Project Foundation** *(this phase)* — repo scaffolding, health check, no
   business features.
2. **Authentication & User Foundation** — users, roles, login, JWT, refresh
   tokens, protected routes.
3. **Practicum & Placement** — institutions, agencies, students, supervisors,
   practicums, placements, assignments.
4. **Student Fieldwork** — dashboard, field activities, attendance, hours,
   weekly reports.
5. **Supervisor Workflows** — faculty/agency dashboards, review, verification,
   supervision records, feedback, evaluations.
6. **Competency System** — configurable framework, criteria, rubrics,
   progress tracking (framework itself is an open product decision — see §8).
7. **Resources & Notifications** — resource library, S3 storage, push
   notifications, missing-record alerts.
8. **Reporting & Monitoring** — student/practicum/competency/supervisor
   reports, verification-issue monitoring.
9. **Administrator Web Dashboard** — Next.js app on the same Go API.
