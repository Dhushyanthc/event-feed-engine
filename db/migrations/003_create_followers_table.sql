CREATE TABLE IF NOT EXISTS followers ( 
  follower_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
  following_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMP DEFAULT NOW(),
  PRIMARY KEY (follower_id, following_id)
);

CREATE INDEX IF NOT EXISTS idx_following_id ON followers(following_id);
