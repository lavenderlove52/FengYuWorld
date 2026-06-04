-- +goose Up
ALTER TABLE user_themes
ADD COLUMN hero_config JSONB NOT NULL DEFAULT '{
  "enabled": false,
  "image_ids": [],
  "transition": "fade",
  "interval_ms": 6000,
  "overlay_opacity": 0.4,
  "hero_height_vh": 100,
  "page_width_px": 1280,
  "article_list_width_px": 960,
  "article_layout": "list"
}'::jsonb;

CREATE TABLE home_theme_images (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  upload_id BIGINT NOT NULL UNIQUE REFERENCES uploads(id) ON DELETE RESTRICT,
  url VARCHAR(500) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_home_theme_images_user_created
ON home_theme_images(user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS home_theme_images;
ALTER TABLE user_themes DROP COLUMN IF EXISTS hero_config;
