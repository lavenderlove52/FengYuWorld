-- +goose Up
ALTER TABLE home_theme_images
  ADD COLUMN mobile_upload_id BIGINT UNIQUE REFERENCES uploads(id) ON DELETE RESTRICT,
  ADD COLUMN mobile_url VARCHAR(500) NOT NULL DEFAULT '',
  ADD CONSTRAINT home_theme_images_mobile_variant_check CHECK (
    (mobile_upload_id IS NULL AND mobile_url = '')
    OR (mobile_upload_id IS NOT NULL AND mobile_url <> '')
  );

-- +goose Down
ALTER TABLE home_theme_images
  DROP CONSTRAINT IF EXISTS home_theme_images_mobile_variant_check,
  DROP COLUMN IF EXISTS mobile_url,
  DROP COLUMN IF EXISTS mobile_upload_id;
