-- +goose Up
CREATE TABLE messages (
  id         BIGSERIAL PRIMARY KEY,
  from_user  BIGINT NOT NULL REFERENCES users(id),
  to_user    BIGINT NOT NULL REFERENCES users(id),
  content    TEXT NOT NULL,
  type       VARCHAR(20) NOT NULL DEFAULT 'text',
  is_read    BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_msg_conversation ON messages (
  LEAST(from_user, to_user),
  GREATEST(from_user, to_user),
  id DESC
);

CREATE INDEX idx_msg_unread ON messages (to_user, is_read) WHERE is_read = false;

-- +goose Down
DROP TABLE IF EXISTS messages;
