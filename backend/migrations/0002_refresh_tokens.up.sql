-- Refresh tokens are stored hashed (never the raw token) and rotated on
-- every use: refreshing revokes the old row and inserts a new one. This
-- lets a compromised refresh token be invalidated without touching the
-- user's password, and lets us revoke all sessions for a user on demand.

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
