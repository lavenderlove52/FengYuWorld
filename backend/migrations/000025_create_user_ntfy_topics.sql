-- +goose Up
CREATE TABLE user_ntfy_topics (
  id             BIGSERIAL    PRIMARY KEY,
  user_id        BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  topic          VARCHAR(80)  NOT NULL UNIQUE,
  public_enabled BOOLEAN      NOT NULL DEFAULT false,
  public_token   VARCHAR(100) NOT NULL DEFAULT '',
  created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX user_ntfy_topics_user_id_idx ON user_ntfy_topics (user_id, created_at);
CREATE UNIQUE INDEX user_ntfy_topics_public_token_uidx ON user_ntfy_topics (public_token) WHERE public_token <> '';

INSERT INTO user_ntfy_topics (user_id, topic, public_enabled, public_token, updated_at)
SELECT user_id, ntfy_topic, ntfy_public_enabled, ntfy_public_token, updated_at
FROM user_push_settings
WHERE ntfy_topic <> ''
ON CONFLICT (topic) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS user_ntfy_topics;
