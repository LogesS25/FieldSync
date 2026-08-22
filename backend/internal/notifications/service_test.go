package notifications_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/notifications"
	"github.com/fieldsync/backend/internal/testutil"
)

func TestCreateAndListForUser(t *testing.T) {
	queries := testutil.NewTestQueries(t)
	svc := notifications.NewService(queries)
	ctx := context.Background()
	user := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	if _, err := svc.Create(ctx, user.ID, "first"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := svc.Create(ctx, user.ID, "second"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	list, err := svc.ListForUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListForUser: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len(list) = %d, want 2", len(list))
	}
	if list[0].Message != "second" {
		t.Errorf("list[0].Message = %q, want most recent first", list[0].Message)
	}
	if list[0].ReadAt.Valid {
		t.Errorf("new notification should be unread")
	}
}

func TestMarkRead_RejectsOtherUser(t *testing.T) {
	queries := testutil.NewTestQueries(t)
	svc := notifications.NewService(queries)
	ctx := context.Background()
	user := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	other := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	n, err := svc.Create(ctx, user.ID, "hello")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = svc.MarkRead(ctx, n.ID, other.ID)
	if !errors.Is(err, notifications.ErrNotYourNotification) {
		t.Fatalf("error = %v, want ErrNotYourNotification", err)
	}
}

func TestMarkRead_Success(t *testing.T) {
	queries := testutil.NewTestQueries(t)
	svc := notifications.NewService(queries)
	ctx := context.Background()
	user := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	n, err := svc.Create(ctx, user.ID, "hello")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated, err := svc.MarkRead(ctx, n.ID, user.ID)
	if err != nil {
		t.Fatalf("MarkRead: %v", err)
	}
	if !updated.ReadAt.Valid {
		t.Errorf("ReadAt should be set after MarkRead")
	}
}

func TestRegisterPushToken_UpsertReassignsToNewUser(t *testing.T) {
	queries := testutil.NewTestQueries(t)
	svc := notifications.NewService(queries)
	ctx := context.Background()
	userA := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	userB := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	if _, err := svc.RegisterPushToken(ctx, userA.ID, "device-token-1"); err != nil {
		t.Fatalf("RegisterPushToken(userA): %v", err)
	}
	// Same device, different user (e.g. someone logged out and a different
	// user logged in) — should reassign, not error.
	if _, err := svc.RegisterPushToken(ctx, userB.ID, "device-token-1"); err != nil {
		t.Fatalf("RegisterPushToken(userB): %v", err)
	}

	tokens, err := queries.ListPushTokensForUser(ctx, userB.ID)
	if err != nil {
		t.Fatalf("ListPushTokensForUser: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("len(tokens) for userB = %d, want 1", len(tokens))
	}

	tokensA, err := queries.ListPushTokensForUser(ctx, userA.ID)
	if err != nil {
		t.Fatalf("ListPushTokensForUser: %v", err)
	}
	if len(tokensA) != 0 {
		t.Fatalf("len(tokens) for userA = %d, want 0 (reassigned away)", len(tokensA))
	}
}

func TestUnregisterPushToken(t *testing.T) {
	queries := testutil.NewTestQueries(t)
	svc := notifications.NewService(queries)
	ctx := context.Background()
	user := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	if _, err := svc.RegisterPushToken(ctx, user.ID, "device-token-2"); err != nil {
		t.Fatalf("RegisterPushToken: %v", err)
	}
	if err := svc.UnregisterPushToken(ctx, "device-token-2"); err != nil {
		t.Fatalf("UnregisterPushToken: %v", err)
	}

	tokens, err := queries.ListPushTokensForUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListPushTokensForUser: %v", err)
	}
	if len(tokens) != 0 {
		t.Fatalf("len(tokens) = %d, want 0 after unregister", len(tokens))
	}
}

func TestCreate_NoPushTokensRegisteredIsNotAnError(t *testing.T) {
	queries := testutil.NewTestQueries(t)
	svc := notifications.NewService(queries)
	ctx := context.Background()
	user := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	// No push token registered for this user — Create must still succeed
	// (the push dispatch is a silent no-op, not an error).
	if _, err := svc.Create(ctx, user.ID, "hello"); err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestMarkAllRead(t *testing.T) {
	queries := testutil.NewTestQueries(t)
	svc := notifications.NewService(queries)
	ctx := context.Background()
	user := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	if _, err := svc.Create(ctx, user.ID, "a"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := svc.Create(ctx, user.ID, "b"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.MarkAllRead(ctx, user.ID); err != nil {
		t.Fatalf("MarkAllRead: %v", err)
	}

	list, err := svc.ListForUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListForUser: %v", err)
	}
	for _, n := range list {
		if !n.ReadAt.Valid {
			t.Errorf("notification %s still unread after MarkAllRead", n.ID)
		}
	}
}
