-- +goose Up
CREATE TABLE user_push_settings (
  user_id       BIGINT        PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  bark_endpoint VARCHAR(1000) NOT NULL DEFAULT '',
  updated_at    TIMESTAMPTZ   NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS user_push_settings;
