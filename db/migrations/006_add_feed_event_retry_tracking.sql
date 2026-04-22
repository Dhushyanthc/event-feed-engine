ALTER TABLE feed_events
ADD COLUMN IF NOT EXISTS attempts INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS failed BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS last_error TEXT;

CREATE INDEX IF NOT EXISTS idx_feed_events_retryable
ON feed_events(processed, failed, processing, attempts, created_at);
