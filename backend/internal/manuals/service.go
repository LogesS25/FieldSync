// Package manuals implements the Guidelines & Manuals feature (business
// requirements §17): a university uploads a PDF guidance document, visible
// to every stakeholder at that university (students, faculty supervisors,
// and agency supervisors at agencies belonging to it). Manual
// versioning/archiving/per-user-visibility rules are explicitly TBD in the
// requirements doc, so this deliberately keeps to what IS specified: one
// current manual per university, replaced wholesale on re-upload.
package manuals

import (
	"context"
	"errors"
	"io"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/storage"
)

var (
	ErrManualNotFound     = errors.New("manual not found")
	ErrNotYourInstitution = errors.New("this manual does not belong to your institution")
)

const storageSubdir = "manuals"

type Service struct {
	queries *sqlcgen.Queries
	storage *storage.Storage
}

func NewService(queries *sqlcgen.Queries, store *storage.Storage) *Service {
	return &Service{queries: queries, storage: store}
}

func (s *Service) Upload(ctx context.Context, institutionID, uploaderID pgtype.UUID, originalFilename string, file io.Reader) (sqlcgen.Manual, error) {
	filePath, err := s.storage.Save(storageSubdir, ".pdf", file)
	if err != nil {
		return sqlcgen.Manual{}, err
	}

	return s.queries.UpsertManual(ctx, sqlcgen.UpsertManualParams{
		InstitutionID:    institutionID,
		FilePath:         filePath,
		OriginalFilename: originalFilename,
		UploadedBy:       uploaderID,
	})
}

func (s *Service) List(ctx context.Context) ([]sqlcgen.Manual, error) {
	return s.queries.ListManuals(ctx)
}

func (s *Service) Delete(ctx context.Context, institutionID pgtype.UUID) error {
	return s.queries.DeleteManual(ctx, institutionID)
}

func (s *Service) GetForUser(ctx context.Context, callerID pgtype.UUID) (sqlcgen.Manual, error) {
	manual, err := s.queries.GetManualForUser(ctx, callerID)
	if err != nil {
		return sqlcgen.Manual{}, ErrManualNotFound
	}
	return manual, nil
}

// GetFileForDownload returns the absolute on-disk path and original
// filename for a manual, after checking the caller belongs to that
// manual's university (administrators may download any manual).
func (s *Service) GetFileForDownload(ctx context.Context, manualID, callerID pgtype.UUID, callerIsAdmin bool) (absolutePath, filename string, err error) {
	manual, getErr := s.queries.GetManualByID(ctx, manualID)
	if getErr != nil {
		return "", "", ErrManualNotFound
	}

	if !callerIsAdmin {
		effectiveInstitution, effErr := s.queries.GetEffectiveInstitutionForUser(ctx, callerID)
		if effErr != nil || !effectiveInstitution.Valid || effectiveInstitution != manual.InstitutionID {
			return "", "", ErrNotYourInstitution
		}
	}

	return s.storage.AbsolutePath(manual.FilePath), manual.OriginalFilename, nil
}
