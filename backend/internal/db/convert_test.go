package db

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestParseDate_Valid(t *testing.T) {
	d, err := ParseDate("2026-03-15")
	if err != nil {
		t.Fatalf("ParseDate returned error: %v", err)
	}
	if !d.Valid {
		t.Fatal("expected Valid = true")
	}
	if got := DateToStringPtr(d); got == nil || *got != "2026-03-15" {
		t.Errorf("round trip = %v, want 2026-03-15", got)
	}
}

func TestParseDate_Invalid(t *testing.T) {
	_, err := ParseDate("not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid date string")
	}
}

func TestParseDate_WrongFormat(t *testing.T) {
	_, err := ParseDate("03/15/2026")
	if err == nil {
		t.Fatal("expected error for non-ISO date format")
	}
}

func TestParseOptionalDate_Empty(t *testing.T) {
	d, err := ParseOptionalDate("")
	if err != nil {
		t.Fatalf("ParseOptionalDate returned error: %v", err)
	}
	if d.Valid {
		t.Fatal("expected an empty string to produce an invalid/NULL date, not an error")
	}
}

func TestParseOptionalDate_NonEmpty(t *testing.T) {
	d, err := ParseOptionalDate("2026-01-01")
	if err != nil {
		t.Fatalf("ParseOptionalDate returned error: %v", err)
	}
	if !d.Valid {
		t.Fatal("expected Valid = true for a non-empty date string")
	}
}

func TestDateToStringPtr_Invalid(t *testing.T) {
	if got := DateToStringPtr(pgtype.Date{}); got != nil {
		t.Errorf("DateToStringPtr(invalid) = %v, want nil", got)
	}
}

func TestUUIDToStringPtr_Invalid(t *testing.T) {
	if got := UUIDToStringPtr(pgtype.UUID{}); got != nil {
		t.Errorf("UUIDToStringPtr(invalid) = %v, want nil", got)
	}
}

func TestUUIDToStringPtr_Valid(t *testing.T) {
	id, err := ParseUUID("00000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatalf("ParseUUID returned error: %v", err)
	}
	got := UUIDToStringPtr(id)
	if got == nil || *got != "00000000-0000-0000-0000-000000000001" {
		t.Errorf("UUIDToStringPtr = %v, want 00000000-0000-0000-0000-000000000001", got)
	}
}

func TestTextToStringPtr_Invalid(t *testing.T) {
	if got := TextToStringPtr(pgtype.Text{}); got != nil {
		t.Errorf("TextToStringPtr(invalid) = %v, want nil", got)
	}
}

func TestTextToStringPtr_Valid(t *testing.T) {
	got := TextToStringPtr(pgtype.Text{String: "hello", Valid: true})
	if got == nil || *got != "hello" {
		t.Errorf("TextToStringPtr = %v, want hello", got)
	}
}
