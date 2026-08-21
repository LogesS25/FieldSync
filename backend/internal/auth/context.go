package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db"
)

var ErrInvalidUserInToken = errors.New("invalid user in token")

// CurrentUserID reads the authenticated user's ID set by RequireAuth. Only
// valid to call on a route behind RequireAuth — that invariant is enforced
// by construction, not checked here, since every route that needs the
// caller's identity is already wrapped in RequireAuth.
func CurrentUserID(c *gin.Context) (pgtype.UUID, error) {
	idStr, _ := c.Get(ContextUserIDKey)
	id, err := db.ParseUUID(idStr.(string))
	if err != nil {
		return pgtype.UUID{}, ErrInvalidUserInToken
	}
	return id, nil
}

// CurrentUserRole reads the authenticated user's role set by RequireAuth.
// Only valid to call on a route behind RequireAuth.
func CurrentUserRole(c *gin.Context) (string, error) {
	role, ok := c.Get(ContextRoleKey)
	if !ok {
		return "", ErrInvalidUserInToken
	}
	roleStr, ok := role.(string)
	if !ok {
		return "", ErrInvalidUserInToken
	}
	return roleStr, nil
}
