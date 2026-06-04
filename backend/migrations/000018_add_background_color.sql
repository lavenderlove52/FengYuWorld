-- +goose Up
ALTER TABLE user_themes ADD COLUMN background_color VARCHAR(20) NOT NULL DEFAULT '';
ALTER TABLE user_themes ADD COLUMN background_color_enabled BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE user_themes DROP COLUMN IF EXISTS background_color_enabled;
ALTER TABLE user_themes DROP COLUMN IF EXISTS background_color;
