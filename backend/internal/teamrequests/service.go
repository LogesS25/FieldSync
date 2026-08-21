// Package teamrequests implements student-initiated practicum team
// formation: a student names an agency, a faculty supervisor, and an agency
// supervisor; each supervisor independently accepts or rejects; the
// practicum team (Practicum + Placement + both SupervisorAssignments) forms
// automatically once both have accepted. Replaces the old admin-unilateral
// assignment flow — see docs/ARCHITECTURE.md §3a #1.
package teamrequests

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/notifications"
	"github.com/fieldsync/backend/internal/practicums"
)

var (
	ErrStudentNotFound                   = errors.New("student not found")
	ErrUserIsNotStudent                  = errors.New("target user is not a student")
	ErrStudentHasNoUniversity            = errors.New("student is not associated with a university yet")
	ErrAgencyNotFound                    = errors.New("agency not found")
	ErrAgencyWrongUniversity             = errors.New("agency does not belong to the student's university")
	ErrFacultyNotFound                   = errors.New("faculty supervisor not found")
	ErrFacultyWrongRole                  = errors.New("target user is not a faculty supervisor")
	ErrFacultyWrongUniversity            = errors.New("faculty supervisor does not belong to the student's university")
	ErrAgencySupNotFound                 = errors.New("agency supervisor not found")
	ErrAgencySupWrongRole                = errors.New("target user is not an agency supervisor")
	ErrAgencySupWrongAgency              = errors.New("agency supervisor does not belong to the selected agency")
	ErrFieldworkComponentNotFound        = errors.New("fieldwork component not found")
	ErrFieldworkComponentWrongUniversity = errors.New("fieldwork component does not belong to the student's university")
	ErrRequestNotFound                   = errors.New("team request not found")
	ErrNotYourRequest                    = errors.New("this request was not sent to you")
	ErrAlreadyDecided                    = errors.New("you have already responded to this request")
)

type Service struct {
	queries       *sqlcgen.Queries
	practicums    *practicums.Service
	notifications *notifications.Service
}

func NewService(queries *sqlcgen.Queries, practicumsService *practicums.Service, notificationsService *notifications.Service) *Service {
	return &Service{queries: queries, practicums: practicumsService, notifications: notificationsService}
}

// notify is best-effort: a failed notification insert must never fail the
// business action that triggered it.
func (s *Service) notify(ctx context.Context, recipientID pgtype.UUID, message string) {
	if _, err := s.notifications.Create(ctx, recipientID, message); err != nil {
		log.Printf("teamrequests: failed to create notification for %s: %v", recipientID, err)
	}
}

type CreateRequestInput struct {
	StudentID            pgtype.UUID
	AgencyID             pgtype.UUID
	FacultySupervisorID  pgtype.UUID
	AgencySupervisorID   pgtype.UUID
	FieldworkComponentID pgtype.UUID
	FieldworkDescription string
	StartDate            pgtype.Date
}

func (s *Service) CreateRequest(ctx context.Context, in CreateRequestInput) (sqlcgen.PracticumTeamRequest, error) {
	student, err := s.queries.GetUserByID(ctx, in.StudentID)
	if err != nil {
		return sqlcgen.PracticumTeamRequest{}, ErrStudentNotFound
	}
	if student.Role != sqlcgen.UserRoleStudent {
		return sqlcgen.PracticumTeamRequest{}, ErrUserIsNotStudent
	}
	if !student.InstitutionID.Valid {
		return sqlcgen.PracticumTeamRequest{}, ErrStudentHasNoUniversity
	}

	agency, err := s.queries.GetAgencyByID(ctx, in.AgencyID)
	if err != nil {
		return sqlcgen.PracticumTeamRequest{}, ErrAgencyNotFound
	}
	if agency.InstitutionID != student.InstitutionID {
		return sqlcgen.PracticumTeamRequest{}, ErrAgencyWrongUniversity
	}

	faculty, err := s.queries.GetUserByID(ctx, in.FacultySupervisorID)
	if err != nil {
		return sqlcgen.PracticumTeamRequest{}, ErrFacultyNotFound
	}
	if faculty.Role != sqlcgen.UserRoleFacultySupervisor {
		return sqlcgen.PracticumTeamRequest{}, ErrFacultyWrongRole
	}
	if faculty.InstitutionID != student.InstitutionID {
		return sqlcgen.PracticumTeamRequest{}, ErrFacultyWrongUniversity
	}

	agencySup, err := s.queries.GetUserByID(ctx, in.AgencySupervisorID)
	if err != nil {
		return sqlcgen.PracticumTeamRequest{}, ErrAgencySupNotFound
	}
	if agencySup.Role != sqlcgen.UserRoleAgencySupervisor {
		return sqlcgen.PracticumTeamRequest{}, ErrAgencySupWrongRole
	}
	if agencySup.AgencyID != in.AgencyID {
		return sqlcgen.PracticumTeamRequest{}, ErrAgencySupWrongAgency
	}

	component, err := s.queries.GetFieldworkComponentByID(ctx, in.FieldworkComponentID)
	if err != nil {
		return sqlcgen.PracticumTeamRequest{}, ErrFieldworkComponentNotFound
	}
	if component.InstitutionID != student.InstitutionID {
		return sqlcgen.PracticumTeamRequest{}, ErrFieldworkComponentWrongUniversity
	}

	request, err := s.queries.CreateTeamRequest(ctx, sqlcgen.CreateTeamRequestParams{
		StudentID:            in.StudentID,
		InstitutionID:        student.InstitutionID,
		AgencyID:             in.AgencyID,
		FacultySupervisorID:  in.FacultySupervisorID,
		AgencySupervisorID:   in.AgencySupervisorID,
		FieldworkComponentID: in.FieldworkComponentID,
		FieldworkDescription: in.FieldworkDescription,
		StartDate:            in.StartDate,
	})
	if err != nil {
		return sqlcgen.PracticumTeamRequest{}, err
	}

	// Business requirements §8: "the supervisors must receive a
	// notification" when a student sends a team request.
	s.notify(ctx, in.FacultySupervisorID, "You have a new practicum team request awaiting your response.")
	s.notify(ctx, in.AgencySupervisorID, "You have a new practicum team request awaiting your response.")

	return request, nil
}

func (s *Service) ListForStudent(ctx context.Context, studentID pgtype.UUID) ([]sqlcgen.PracticumTeamRequest, error) {
	return s.queries.ListTeamRequestsForStudent(ctx, studentID)
}

func (s *Service) ListForSupervisor(ctx context.Context, supervisorID pgtype.UUID) ([]sqlcgen.PracticumTeamRequest, error) {
	return s.queries.ListTeamRequestsForSupervisor(ctx, supervisorID)
}

// RespondAsFaculty and RespondAsAgency are separate methods (rather than one
// generic "respond" taking a role) because the ownership check and the
// column being written genuinely differ per side — collapsing them would
// need a role-keyed branch internally anyway.
func (s *Service) RespondAsFaculty(ctx context.Context, requestID, supervisorID pgtype.UUID, accept bool) (sqlcgen.PracticumTeamRequest, error) {
	request, err := s.queries.GetTeamRequestByID(ctx, requestID)
	if err != nil {
		return sqlcgen.PracticumTeamRequest{}, ErrRequestNotFound
	}
	if request.FacultySupervisorID != supervisorID {
		return sqlcgen.PracticumTeamRequest{}, ErrNotYourRequest
	}
	if request.FacultyDecision != sqlcgen.TeamRequestDecisionPending {
		return sqlcgen.PracticumTeamRequest{}, ErrAlreadyDecided
	}

	decision := sqlcgen.TeamRequestDecisionRejected
	if accept {
		decision = sqlcgen.TeamRequestDecisionAccepted
	}

	updated, err := s.queries.SetFacultyDecision(ctx, sqlcgen.SetFacultyDecisionParams{
		ID:              requestID,
		FacultyDecision: decision,
	})
	if err != nil {
		return sqlcgen.PracticumTeamRequest{}, err
	}

	s.notify(ctx, updated.StudentID, "Your faculty supervisor "+decisionVerb(accept)+" your team request.")
	return s.maybeFormTeam(ctx, updated)
}

func (s *Service) RespondAsAgency(ctx context.Context, requestID, supervisorID pgtype.UUID, accept bool) (sqlcgen.PracticumTeamRequest, error) {
	request, err := s.queries.GetTeamRequestByID(ctx, requestID)
	if err != nil {
		return sqlcgen.PracticumTeamRequest{}, ErrRequestNotFound
	}
	if request.AgencySupervisorID != supervisorID {
		return sqlcgen.PracticumTeamRequest{}, ErrNotYourRequest
	}
	if request.AgencyDecision != sqlcgen.TeamRequestDecisionPending {
		return sqlcgen.PracticumTeamRequest{}, ErrAlreadyDecided
	}

	decision := sqlcgen.TeamRequestDecisionRejected
	if accept {
		decision = sqlcgen.TeamRequestDecisionAccepted
	}

	updated, err := s.queries.SetAgencyDecision(ctx, sqlcgen.SetAgencyDecisionParams{
		ID:             requestID,
		AgencyDecision: decision,
	})
	if err != nil {
		return sqlcgen.PracticumTeamRequest{}, err
	}

	s.notify(ctx, updated.StudentID, "Your agency supervisor "+decisionVerb(accept)+" your team request.")
	return s.maybeFormTeam(ctx, updated)
}

func decisionVerb(accept bool) string {
	if accept {
		return "accepted"
	}
	return "rejected"
}

// maybeFormTeam creates the Practicum, Placement, and both
// SupervisorAssignments the moment both decisions are 'accepted' — the
// business requirements' diagram (§8) treats "both accept" as the trigger
// for the practicum team to exist.
func (s *Service) maybeFormTeam(ctx context.Context, request sqlcgen.PracticumTeamRequest) (sqlcgen.PracticumTeamRequest, error) {
	if request.FacultyDecision != sqlcgen.TeamRequestDecisionAccepted ||
		request.AgencyDecision != sqlcgen.TeamRequestDecisionAccepted {
		return request, nil
	}

	practicum, err := s.practicums.CreatePracticum(ctx, request.StudentID, request.InstitutionID, request.StartDate, pgtype.Date{})
	if err != nil {
		return request, err
	}
	if _, err := s.practicums.CreatePlacement(ctx, practicum.ID, request.AgencyID, request.StartDate, pgtype.Date{}); err != nil {
		return request, err
	}
	if _, err := s.practicums.CreateSupervisorAssignment(ctx, practicum.ID, request.FacultySupervisorID); err != nil {
		return request, err
	}
	if _, err := s.practicums.CreateSupervisorAssignment(ctx, practicum.ID, request.AgencySupervisorID); err != nil {
		return request, err
	}

	if err := s.queries.MarkTeamRequestFormed(ctx, sqlcgen.MarkTeamRequestFormedParams{
		ID:                request.ID,
		FormedPracticumID: practicum.ID,
	}); err != nil {
		return request, err
	}

	request.FormedPracticumID = practicum.ID

	s.notify(ctx, request.StudentID, "Your practicum team has been formed.")
	s.notify(ctx, request.FacultySupervisorID, "The practicum team has been formed.")
	s.notify(ctx, request.AgencySupervisorID, "The practicum team has been formed.")

	return request, nil
}
