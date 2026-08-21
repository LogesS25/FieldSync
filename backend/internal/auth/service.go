package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
)

var (
	ErrEmailTaken          = errors.New("email already registered")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrInstitutionRequired = errors.New("institutionId is required for this role")
	ErrAgencyRequired      = errors.New("agencyId is required for this role")
	ErrInstitutionNotFound = errors.New("institution not found")
	ErrAgencyNotFound      = errors.New("agency not found")
)

// RegisterableRoles are the roles a user may self-register as. Administrator
// accounts are provisioned separately (Phase 9 Admin dashboard) and are
// deliberately excluded here.
var RegisterableRoles = map[sqlcgen.UserRole]bool{
	sqlcgen.UserRoleStudent:           true,
	sqlcgen.UserRoleFacultySupervisor: true,
	sqlcgen.UserRoleAgencySupervisor:  true,
}

type Session struct {
	User         sqlcgen.User
	AccessToken  string
	RefreshToken string
}

type Service struct {
	queries         *sqlcgen.Queries
	jwtSecret       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewService(queries *sqlcgen.Queries, jwtSecret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		queries:         queries,
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

// RegisterInput's InstitutionID/AgencyID are required per-role: student and
// faculty_supervisor need InstitutionID, agency_supervisor needs AgencyID —
// business requirements §3 ties every stakeholder to a university, directly
// or (for agency supervisors) via their agency. Without this, a student
// couldn't send a practicum team request (internal/teamrequests validates
// the agency/faculty they pick belong to their own university).
type RegisterInput struct {
	Email         string
	Password      string
	FullName      string
	Role          sqlcgen.UserRole
	InstitutionID pgtype.UUID
	AgencyID      pgtype.UUID
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (Session, error) {
	switch in.Role {
	case sqlcgen.UserRoleStudent, sqlcgen.UserRoleFacultySupervisor:
		if !in.InstitutionID.Valid {
			return Session{}, ErrInstitutionRequired
		}
		if _, err := s.queries.GetInstitutionByID(ctx, in.InstitutionID); err != nil {
			return Session{}, ErrInstitutionNotFound
		}
	case sqlcgen.UserRoleAgencySupervisor:
		if !in.AgencyID.Valid {
			return Session{}, ErrAgencyRequired
		}
		if _, err := s.queries.GetAgencyByID(ctx, in.AgencyID); err != nil {
			return Session{}, ErrAgencyNotFound
		}
	}

	passwordHash, err := HashPassword(in.Password)
	if err != nil {
		return Session{}, err
	}

	user, err := s.queries.CreateUser(ctx, sqlcgen.CreateUserParams{
		Email:         in.Email,
		PasswordHash:  passwordHash,
		Role:          in.Role,
		FullName:      in.FullName,
		InstitutionID: in.InstitutionID,
		AgencyID:      in.AgencyID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Session{}, ErrEmailTaken
		}
		return Session{}, err
	}

	return s.issueSession(ctx, user)
}

func (s *Service) Login(ctx context.Context, email, password string) (Session, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return Session{}, ErrInvalidCredentials
	}

	if !CheckPassword(password, user.PasswordHash) {
		return Session{}, ErrInvalidCredentials
	}

	return s.issueSession(ctx, user)
}

// Refresh rotates the refresh token: the presented token is revoked and a
// new access/refresh pair is issued, so a leaked refresh token only works
// once before falling out of sync with the legitimate client.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (Session, error) {
	record, err := s.queries.GetActiveRefreshToken(ctx, hashToken(refreshToken))
	if err != nil {
		return Session{}, ErrInvalidRefreshToken
	}

	user, err := s.queries.GetUserByID(ctx, record.UserID)
	if err != nil {
		return Session{}, ErrInvalidRefreshToken
	}

	if err := s.queries.RevokeRefreshToken(ctx, record.ID); err != nil {
		return Session{}, err
	}

	return s.issueSession(ctx, user)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	record, err := s.queries.GetActiveRefreshToken(ctx, hashToken(refreshToken))
	if err != nil {
		// Already invalid/expired/revoked — logout is idempotent.
		return nil
	}
	return s.queries.RevokeRefreshToken(ctx, record.ID)
}

func (s *Service) issueSession(ctx context.Context, user sqlcgen.User) (Session, error) {
	accessToken, err := GenerateAccessToken(s.jwtSecret, db.UUIDToString(user.ID), string(user.Role), s.accessTokenTTL)
	if err != nil {
		return Session{}, err
	}

	refreshToken, err := generateOpaqueToken()
	if err != nil {
		return Session{}, err
	}

	_, err = s.queries.CreateRefreshToken(ctx, sqlcgen.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(s.refreshTokenTTL), Valid: true},
	})
	if err != nil {
		return Session{}, err
	}

	return Session{User: user, AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
