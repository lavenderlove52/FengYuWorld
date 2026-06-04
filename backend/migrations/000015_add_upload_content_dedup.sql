-- +goose Up
CREATE TABLE upload_assets (
  id BIGSERIAL PRIMARY KEY,
  content_hash CHAR(64) NOT NULL,
  variant VARCHAR(50) NOT NULL,
  path VARCHAR(500) NOT NULL,
  url VARCHAR(500) NOT NULL,
  mime VARCHAR(120) NOT NULL,
  size_bytes BIGINT NOT NULL,
  original_size_bytes BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT upload_assets_content_hash_variant_key UNIQUE (content_hash, variant),
  CONSTRAINT upload_assets_size_check CHECK (size_bytes > 0 AND original_size_bytes > 0)
);

ALTER TABLE uploads
  ADD COLUMN asset_id BIGINT REFERENCES upload_assets(id),
  ADD COLUMN purpose VARCHAR(50),
  ADD COLUMN owner_type VARCHAR(50),
  ADD COLUMN owner_key VARCHAR(200),
  ADD CONSTRAINT uploads_asset_reference_fields_check CHECK (
    asset_id IS NULL OR (purpose IS NOT NULL AND owner_type IS NOT NULL AND owner_key IS NOT NULL)
  );

CREATE UNIQUE INDEX uploads_logical_reference_key
  ON uploads (asset_id, user_id, purpose, owner_type, owner_key)
  WHERE asset_id IS NOT NULL;

CREATE INDEX uploads_asset_id_idx
  ON uploads (asset_id)
  WHERE asset_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS uploads_asset_id_idx;
DROP INDEX IF EXISTS uploads_logical_reference_key;

ALTER TABLE uploads
  DROP CONSTRAINT IF EXISTS uploads_asset_reference_fields_check,
  DROP COLUMN IF EXISTS owner_key,
  DROP COLUMN IF EXISTS owner_type,
  DROP COLUMN IF EXISTS purpose,
  DROP COLUMN IF EXISTS asset_id;

DROP TABLE IF EXISTS upload_assets;
