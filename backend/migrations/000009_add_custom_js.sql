-- +goose Up
ALTER TABLE users ADD COLUMN can_custom_js BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE user_themes ADD COLUMN custom_js TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE user_themes DROP COLUMN IF EXISTS custom_js;
ALTER TABLE users DROP COLUMN IF EXISTS can_custom_js;
