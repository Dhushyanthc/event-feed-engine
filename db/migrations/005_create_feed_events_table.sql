CREATE TABLE IF NOT EXISTS feed_events (
  id BIGSERIAL PRIMARY KEY,
  post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  processed BOOLEAN NOT NULL DEFAULT FALSE,
  processing BOOLEAN NOT NULL DEFAULT FALSE,
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feed_events_unprocessed
ON feed_events(processed, processing, created_at);

CREATE INDEX IF NOT EXISTS idx_feed_events_updated_at
ON feed_events(updated_at);

