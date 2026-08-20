// Package attendance implements Student attendance recording with logged
// hours, feeding the "field hours" tracking goal. Like field_activities,
// records are write-once from the student's side (see internal/activities
// package doc for the rationale).
package attendance

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/practicums"
)

type Service struct {
	queries    *sqlcgen.Queries
	practicums *practicums.Service
}

func NewService(queries *sqlcgen.Queries, practicumsService *practicums.Service) *Service {
	return &Service{queries: queries, practicums: practicumsService}
}

func (s *Service) Create(ctx context.Context, studentID pgtype.UUID, attendanceDate pgtype.Date, hours float64) (sqlcgen.AttendanceRecord, error) {
	practicumID, err := s.practicums.GetActivePracticumID(ctx, studentID)
	if err != nil {
		return sqlcgen.AttendanceRecord{}, err
	}

	hoursLogged, err := db.ParseHours(hours)
	if err != nil {
		return sqlcgen.AttendanceRecord{}, err
	}

	return s.queries.CreateAttendanceRecord(ctx, sqlcgen.CreateAttendanceRecordParams{
		StudentID:      studentID,
		PracticumID:    practicumID,
		AttendanceDate: attendanceDate,
		HoursLogged:    hoursLogged,
	})
}

func (s *Service) ListForStudent(ctx context.Context, studentID pgtype.UUID) ([]sqlcgen.AttendanceRecord, error) {
	return s.queries.ListAttendanceForStudent(ctx, studentID)
}

func (s *Service) GetTotalHours(ctx context.Context, studentID pgtype.UUID) (float64, error) {
	total, err := s.queries.GetTotalHoursForStudent(ctx, studentID)
	if err != nil {
		return 0, err
	}
	return db.NumericToFloat64(total), nil
}
