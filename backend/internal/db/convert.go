package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const dateLayout = "2006-01-02"

func ParseDate(s string) (pgtype.Date, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return pgtype.Date{}, err
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

// ParseOptionalDate returns a zero-value (invalid/NULL) pgtype.Date for an
// empty string, since end dates are frequently unset (practicum in
// progress).
func ParseOptionalDate(s string) (pgtype.Date, error) {
	if s == "" {
		return pgtype.Date{}, nil
	}
	return ParseDate(s)
}

func DateToStringPtr(d pgtype.Date) *string {
	if !d.Valid {
		return nil
	}
	s := d.Time.Format(dateLayout)
	return &s
}

func UUIDToStringPtr(id pgtype.UUID) *string {
	if !id.Valid {
		return nil
	}
	s := UUIDToString(id)
	return &s
}

func TextToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}
