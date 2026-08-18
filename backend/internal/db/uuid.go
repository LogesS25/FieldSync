package db

import "github.com/jackc/pgx/v5/pgtype"

func UUIDToString(id pgtype.UUID) string {
	uuidVal, _ := id.Value()
	s, _ := uuidVal.(string)
	return s
}

func ParseUUID(s string) (pgtype.UUID, error) {
	var id pgtype.UUID
	err := id.Scan(s)
	return id, err
}
