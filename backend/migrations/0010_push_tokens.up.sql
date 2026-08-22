-- Device push tokens (Expo push notification service) for delivering
-- notifications (see 0009_notifications) even when the app isn't open.
-- A device re-registering (e.g. after a different user logs in on a
-- shared device) reassigns the token to the new user rather than erroring.

CREATE TABLE push_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX idx_push_tokens_user_id ON push_tokens (user_id);
