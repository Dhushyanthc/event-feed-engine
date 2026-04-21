CREATE TABLE IF NOT EXISTS user_feed (
  user_id BIGINT NOT NULL,
  post_id BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL,

  PRIMARY KEY (user_id, post_id)
);

CREATE INDEX IF NOT EXISTS idx_feed_user_time
ON user_feed(user_id, created_at DESC);
