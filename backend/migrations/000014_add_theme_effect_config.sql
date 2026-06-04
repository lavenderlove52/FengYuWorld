-- +goose Up
ALTER TABLE user_themes
ADD COLUMN effect_config JSONB NOT NULL DEFAULT '{
  "scope": "all",
  "background": {"type": "off", "color": "#6366f1", "count": 20, "size": 20, "opacity": 0.25, "speed": 0.5},
  "cursor": {"type": "off", "color": "#6366f1", "count": 16, "size": 5, "opacity": 0.7, "duration": 700},
  "click": {"type": "off", "color": "#ec4899", "count": 8, "size": 16, "opacity": 0.9, "duration": 900}
}'::jsonb;

-- +goose Down
ALTER TABLE user_themes DROP COLUMN IF EXISTS effect_config;
