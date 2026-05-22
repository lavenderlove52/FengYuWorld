-- +goose Up
CREATE TABLE friendships (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id),
  friend_id BIGINT NOT NULL REFERENCES users(id),
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT friendships_not_self CHECK (user_id <> friend_id),
  CONSTRAINT friendships_unique_pair UNIQUE (user_id, friend_id)
);

CREATE INDEX idx_friendships_user_status ON friendships(user_id, status);
CREATE INDEX idx_friendships_friend_status ON friendships(friend_id, status);
CREATE INDEX idx_friendships_user_friend_status ON friendships(user_id, friend_id, status);

-- +goose Down
DROP TABLE IF EXISTS friendships;
