-- +goose Up
ALTER TABLE user_themes ADD COLUMN theme_color_enabled BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE user_themes DROP COLUMN IF EXISTS theme_color_enabled;
