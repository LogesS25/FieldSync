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
