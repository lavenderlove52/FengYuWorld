-- +goose Up
CREATE TABLE uploads (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id),
  path VARCHAR(500) NOT NULL,
  url VARCHAR(500) NOT NULL,
  mime VARCHAR(120) NOT NULL,
  size_bytes BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS uploads;
