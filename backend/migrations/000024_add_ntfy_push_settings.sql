-- +goose Up
ALTER TABLE user_push_settings
  ADD COLUMN ntfy_topic VARCHAR(80) NOT NULL DEFAULT '',
  ADD COLUMN ntfy_public_enabled BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN ntfy_public_token VARCHAR(100) NOT NULL DEFAULT '';
CREATE UNIQUE INDEX user_push_settings_ntfy_topic_uidx ON user_push_settings (ntfy_topic) WHERE ntfy_topic <> '';
CREATE UNIQUE INDEX user_push_settings_ntfy_public_token_uidx ON user_push_settings (ntfy_public_token) WHERE ntfy_public_token <> '';

-- +goose Down
DROP INDEX IF EXISTS user_push_settings_ntfy_public_token_uidx;
DROP INDEX IF EXISTS user_push_settings_ntfy_topic_uidx;
ALTER TABLE user_push_settings DROP COLUMN IF EXISTS ntfy_public_token, DROP COLUMN IF EXISTS ntfy_public_enabled, DROP COLUMN IF EXISTS ntfy_topic;
