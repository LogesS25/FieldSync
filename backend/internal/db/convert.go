package db

import (
	"strconv"
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

// ParseHours converts a float (e.g. from a JSON request body) into
// pgtype.Numeric via its string representation — pgtype.Numeric doesn't
// accept a float64 directly, and going through the decimal string form
// avoids float64 binary-precision surprises for a value that ends up in a
// NUMERIC(4,2) column anyway.
func ParseHours(hours float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(strconv.FormatFloat(hours, 'f', 2, 64)); err != nil {
		return pgtype.Numeric{}, err
	}
	return n, nil
}

func NumericToFloat64(n pgtype.Numeric) float64 {
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
}
