-- In-app notifications (business requirements §8 "the supervisors must
-- receive a notification" on team request creation, §10 "the agency
-- supervisor and faculty supervisor are notified" on daily report
-- submission). Messages are precomputed at insert time (denormalized) so
-- the mobile client doesn't need per-kind rendering logic.

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    read_at TIMESTAMPTZ,
    -- clock_timestamp(), not now() — now() returns the same value for every
    -- statement within one transaction, which ties multiple notifications
    -- created back-to-back (e.g. team formation notifying three people) and
    -- makes "most recent first" ordering nondeterministic.
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX idx_notifications_recipient_id_created_at ON notifications (recipient_id, created_at DESC);
