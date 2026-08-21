package teamrequests_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/practicums"
	"github.com/fieldsync/backend/internal/teamrequests"
	"github.com/fieldsync/backend/internal/testutil"
)

func parseDate(t *testing.T, s string) pgtype.Date {
	t.Helper()
	d, err := db.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

// createScopedUser inserts a user with institution_id/agency_id set
// directly — mirrors what real registration now requires (see
// internal/auth.RegisterInput). testutil.CreateTestUser doesn't set these
// since most packages don't need them.
func createScopedUser(t *testing.T, queries *sqlcgen.Queries, role sqlcgen.UserRole, institutionID, agencyID pgtype.UUID) sqlcgen.User {
	t.Helper()
	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	user, err := queries.CreateUser(context.Background(), sqlcgen.CreateUserParams{
		Email:         fmt.Sprintf("%s-%d@example.com", role, time.Now().UnixNano()),
		PasswordHash:  hash,
		Role:          role,
		FullName:      "Test " + string(role),
		InstitutionID: institutionID,
		AgencyID:      agencyID,
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	return user
}

type testFixture struct {
	svc         *teamrequests.Service
	queries     *sqlcgen.Queries
	institution sqlcgen.Institution
	agency      sqlcgen.Agency
	student     sqlcgen.User
	faculty     sqlcgen.User
	agencySup   sqlcgen.User
}

func newFixture(t *testing.T) testFixture {
	t.Helper()
	queries := testutil.NewTestQueries(t)
	practicumsSvc := practicums.NewService(queries)

	institution := testutil.CreateTestInstitution(t, queries)
	agency := testutil.CreateTestAgency(t, queries, institution.ID)

	student := createScopedUser(t, queries, sqlcgen.UserRoleStudent, institution.ID, pgtype.UUID{})
	faculty := createScopedUser(t, queries, sqlcgen.UserRoleFacultySupervisor, institution.ID, pgtype.UUID{})
	agencySup := createScopedUser(t, queries, sqlcgen.UserRoleAgencySupervisor, pgtype.UUID{}, agency.ID)

	return testFixture{
		svc:         teamrequests.NewService(queries, practicumsSvc),
		queries:     queries,
		institution: institution,
		agency:      agency,
		student:     student,
		faculty:     faculty,
		agencySup:   agencySup,
	}
}

func TestCreateRequest_Success(t *testing.T) {
	f := newFixture(t)

	request, err := f.svc.CreateRequest(context.Background(), teamrequests.CreateRequestInput{
		StudentID:            f.student.ID,
		AgencyID:             f.agency.ID,
		FacultySupervisorID:  f.faculty.ID,
		AgencySupervisorID:   f.agencySup.ID,
		FieldworkDescription: "Casework at a community clinic.",
		StartDate:            parseDate(t, "2026-01-01"),
	})
	if err != nil {
		t.Fatalf("CreateRequest returned error: %v", err)
	}
	if request.FacultyDecision != sqlcgen.TeamRequestDecisionPending {
		t.Errorf("FacultyDecision = %v, want pending", request.FacultyDecision)
	}
	if request.AgencyDecision != sqlcgen.TeamRequestDecisionPending {
		t.Errorf("AgencyDecision = %v, want pending", request.AgencyDecision)
	}
}

func TestCreateRequest_AgencyFromDifferentUniversityRejected(t *testing.T) {
	f := newFixture(t)
	otherInstitution := testutil.CreateTestInstitution(t, f.queries)
	otherAgency := testutil.CreateTestAgency(t, f.queries, otherInstitution.ID)

	_, err := f.svc.CreateRequest(context.Background(), teamrequests.CreateRequestInput{
		StudentID:            f.student.ID,
		AgencyID:             otherAgency.ID,
		FacultySupervisorID:  f.faculty.ID,
		AgencySupervisorID:   f.agencySup.ID,
		FieldworkDescription: "Should be rejected.",
		StartDate:            parseDate(t, "2026-01-01"),
	})
	if !errors.Is(err, teamrequests.ErrAgencyWrongUniversity) {
		t.Fatalf("error = %v, want ErrAgencyWrongUniversity", err)
	}
}

func TestCreateRequest_AgencySupervisorFromDifferentAgencyRejected(t *testing.T) {
	f := newFixture(t)
	otherAgency := testutil.CreateTestAgency(t, f.queries, f.institution.ID)
	mismatchedSup := createScopedUser(t, f.queries, sqlcgen.UserRoleAgencySupervisor, pgtype.UUID{}, otherAgency.ID)

	_, err := f.svc.CreateRequest(context.Background(), teamrequests.CreateRequestInput{
		StudentID:            f.student.ID,
		AgencyID:             f.agency.ID,
		FacultySupervisorID:  f.faculty.ID,
		AgencySupervisorID:   mismatchedSup.ID,
		FieldworkDescription: "Should be rejected.",
		StartDate:            parseDate(t, "2026-01-01"),
	})
	if !errors.Is(err, teamrequests.ErrAgencySupWrongAgency) {
		t.Fatalf("error = %v, want ErrAgencySupWrongAgency", err)
	}
}

func TestCreateRequest_StudentWithoutUniversityRejected(t *testing.T) {
	f := newFixture(t)
	homelessStudent := createScopedUser(t, f.queries, sqlcgen.UserRoleStudent, pgtype.UUID{}, pgtype.UUID{})

	_, err := f.svc.CreateRequest(context.Background(), teamrequests.CreateRequestInput{
		StudentID:            homelessStudent.ID,
		AgencyID:             f.agency.ID,
		FacultySupervisorID:  f.faculty.ID,
		AgencySupervisorID:   f.agencySup.ID,
		FieldworkDescription: "Should be rejected.",
		StartDate:            parseDate(t, "2026-01-01"),
	})
	if !errors.Is(err, teamrequests.ErrStudentHasNoUniversity) {
		t.Fatalf("error = %v, want ErrStudentHasNoUniversity", err)
	}
}

func TestRespond_BothAcceptFormsTheTeam(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	request, err := f.svc.CreateRequest(ctx, teamrequests.CreateRequestInput{
		StudentID:            f.student.ID,
		AgencyID:             f.agency.ID,
		FacultySupervisorID:  f.faculty.ID,
		AgencySupervisorID:   f.agencySup.ID,
		FieldworkDescription: "Casework.",
		StartDate:            parseDate(t, "2026-01-01"),
	})
	if err != nil {
		t.Fatalf("CreateRequest returned error: %v", err)
	}

	afterFaculty, err := f.svc.RespondAsFaculty(ctx, request.ID, f.faculty.ID, true)
	if err != nil {
		t.Fatalf("RespondAsFaculty returned error: %v", err)
	}
	if afterFaculty.FormedPracticumID.Valid {
		t.Fatal("practicum should not form until both supervisors accept")
	}

	afterAgency, err := f.svc.RespondAsAgency(ctx, request.ID, f.agencySup.ID, true)
	if err != nil {
		t.Fatalf("RespondAsAgency returned error: %v", err)
	}
	if !afterAgency.FormedPracticumID.Valid {
		t.Fatal("expected FormedPracticumID to be set once both supervisors accepted")
	}
}

func TestRespond_OneRejectionDoesNotFormTeam(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	request, err := f.svc.CreateRequest(ctx, teamrequests.CreateRequestInput{
		StudentID:            f.student.ID,
		AgencyID:             f.agency.ID,
		FacultySupervisorID:  f.faculty.ID,
		AgencySupervisorID:   f.agencySup.ID,
		FieldworkDescription: "Casework.",
		StartDate:            parseDate(t, "2026-01-01"),
	})
	if err != nil {
		t.Fatalf("CreateRequest returned error: %v", err)
	}

	if _, err := f.svc.RespondAsFaculty(ctx, request.ID, f.faculty.ID, true); err != nil {
		t.Fatalf("RespondAsFaculty returned error: %v", err)
	}
	afterAgency, err := f.svc.RespondAsAgency(ctx, request.ID, f.agencySup.ID, false)
	if err != nil {
		t.Fatalf("RespondAsAgency returned error: %v", err)
	}
	if afterAgency.FormedPracticumID.Valid {
		t.Fatal("practicum must not form when either supervisor rejects")
	}
}

func TestRespond_WrongSupervisorRejected(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	otherFaculty := createScopedUser(t, f.queries, sqlcgen.UserRoleFacultySupervisor, f.institution.ID, pgtype.UUID{})

	request, err := f.svc.CreateRequest(ctx, teamrequests.CreateRequestInput{
		StudentID:            f.student.ID,
		AgencyID:             f.agency.ID,
		FacultySupervisorID:  f.faculty.ID,
		AgencySupervisorID:   f.agencySup.ID,
		FieldworkDescription: "Casework.",
		StartDate:            parseDate(t, "2026-01-01"),
	})
	if err != nil {
		t.Fatalf("CreateRequest returned error: %v", err)
	}

	_, err = f.svc.RespondAsFaculty(ctx, request.ID, otherFaculty.ID, true)
	if !errors.Is(err, teamrequests.ErrNotYourRequest) {
		t.Fatalf("error = %v, want ErrNotYourRequest", err)
	}
}

func TestRespond_CannotDecideTwice(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	request, err := f.svc.CreateRequest(ctx, teamrequests.CreateRequestInput{
		StudentID:            f.student.ID,
		AgencyID:             f.agency.ID,
		FacultySupervisorID:  f.faculty.ID,
		AgencySupervisorID:   f.agencySup.ID,
		FieldworkDescription: "Casework.",
		StartDate:            parseDate(t, "2026-01-01"),
	})
	if err != nil {
		t.Fatalf("CreateRequest returned error: %v", err)
	}

	if _, err := f.svc.RespondAsFaculty(ctx, request.ID, f.faculty.ID, true); err != nil {
		t.Fatalf("first RespondAsFaculty returned error: %v", err)
	}
	_, err = f.svc.RespondAsFaculty(ctx, request.ID, f.faculty.ID, false)
	if !errors.Is(err, teamrequests.ErrAlreadyDecided) {
		t.Fatalf("error = %v, want ErrAlreadyDecided", err)
	}
}
