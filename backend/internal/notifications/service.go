// Package notifications implements in-app notifications. The business
// requirements explicitly mandate two triggers: supervisors are notified
// when a student sends a team request (§8), and both supervisors are
// notified when a student submits a daily report (§10). Notifying a user
// when their own record is subsequently reviewed is a natural extension of
// the same mechanism (closing the loop), not a new business/scoring rule,
// so it's included too — see the call sites in teamrequests, dailyreports,
// attendance, reports, and feedback.
//
// Messages are precomputed plain text at insert time (denormalized, no
// "kind" enum) so the mobile client needs no per-type rendering logic —
// kept deliberately simple for a first pass; add a kind/icon column later
// if the UI needs to differentiate visually.
//
// Every Create also fires an Expo push notification (see push.go) to any
// device tokens registered for the recipient — fire-and-forget, since a
// push-delivery failure must never block the in-app notification or the
// business action that triggered it.
package notifications

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db/sqlcgen"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrNotYourNotification  = errors.New("this notification does not belong to you")
)

type Service struct {
	queries *sqlcgen.Queries
}

func NewService(queries *sqlcgen.Queries) *Service {
	return &Service{queries: queries}
}

// Create is best-effort from the caller's perspective: a failed
// notification insert should never fail the business action that
// triggered it. Callers log and swallow the error rather than propagating
// it — see call sites.
func (s *Service) Create(ctx context.Context, recipientID pgtype.UUID, message string) (sqlcgen.Notification, error) {
	notification, err := s.queries.CreateNotification(ctx, sqlcgen.CreateNotificationParams{
		RecipientID: recipientID,
		Message:     message,
	})
	if err != nil {
		return sqlcgen.Notification{}, err
	}

	s.dispatchPush(ctx, recipientID, message)

	return notification, nil
}

func (s *Service) ListForUser(ctx context.Context, userID pgtype.UUID) ([]sqlcgen.Notification, error) {
	return s.queries.ListNotificationsForUser(ctx, userID)
}

func (s *Service) MarkRead(ctx context.Context, notificationID, callerID pgtype.UUID) (sqlcgen.Notification, error) {
	notification, err := s.queries.GetNotificationByID(ctx, notificationID)
	if err != nil {
		return sqlcgen.Notification{}, ErrNotificationNotFound
	}
	if notification.RecipientID != callerID {
		return sqlcgen.Notification{}, ErrNotYourNotification
	}
	return s.queries.MarkNotificationRead(ctx, notificationID)
}

func (s *Service) MarkAllRead(ctx context.Context, userID pgtype.UUID) error {
	return s.queries.MarkAllNotificationsRead(ctx, userID)
}
