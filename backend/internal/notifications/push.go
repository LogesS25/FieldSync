package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db/sqlcgen"
)

const expoPushURL = "https://exp.host/--/api/v2/push/send"

type expoPushMessage struct {
	To    string `json:"to"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

// dispatchPush looks up the recipient's registered device tokens (a
// regular query on the caller's connection/transaction — safe, no
// concurrency concerns) and, if any exist, sends the actual push over the
// network in a detached goroutine. The goroutine touches no shared
// *sqlcgen.Queries/transaction state: pgx.Tx (used by the test harness to
// wrap each test in a rolled-back transaction) is not safe for concurrent
// use from multiple goroutines, so all DB access here happens synchronously
// before anything is backgrounded.
func (s *Service) dispatchPush(ctx context.Context, recipientID pgtype.UUID, message string) {
	tokens, err := s.queries.ListPushTokensForUser(ctx, recipientID)
	if err != nil {
		log.Printf("notifications: failed to list push tokens for %s: %v", recipientID, err)
		return
	}
	if len(tokens) == 0 {
		return
	}

	values := make([]string, len(tokens))
	for i, t := range tokens {
		values[i] = t.Token
	}

	go sendExpoPush(values, message)
}

// sendExpoPush is fire-and-forget from the caller's perspective: it runs in
// its own goroutine with its own context (not the request context, which is
// cancelled once the HTTP response is written) so a slow or unreachable
// Expo push endpoint never adds latency to the request that triggered the
// notification. It touches no database state.
func sendExpoPush(tokens []string, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	messages := make([]expoPushMessage, len(tokens))
	for i, token := range tokens {
		messages[i] = expoPushMessage{To: token, Title: "FieldSync", Body: message}
	}

	body, err := json.Marshal(messages)
	if err != nil {
		log.Printf("notifications: failed to marshal push payload: %v", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, expoPushURL, bytes.NewReader(body))
	if err != nil {
		log.Printf("notifications: failed to build push request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("notifications: push request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("notifications: push request returned status %d", resp.StatusCode)
	}
}

func (s *Service) RegisterPushToken(ctx context.Context, userID pgtype.UUID, token string) (sqlcgen.PushToken, error) {
	return s.queries.UpsertPushToken(ctx, sqlcgen.UpsertPushTokenParams{
		UserID: userID,
		Token:  token,
	})
}

func (s *Service) UnregisterPushToken(ctx context.Context, token string) error {
	return s.queries.DeletePushToken(ctx, token)
}
