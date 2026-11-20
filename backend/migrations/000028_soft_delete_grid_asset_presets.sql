-- +goose Up
-- 自建标的改成软删除：用户在下拉里点「移除」后，计划里仍留着同名标的，
-- 下次保存不能被自动收集逻辑又插回来。
ALTER TABLE grid_asset_presets ADD COLUMN deleted_at TIMESTAMPTZ NULL;

CREATE INDEX idx_grid_asset_presets_active ON grid_asset_presets(user_id, market, name)
  WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_grid_asset_presets_active;
ALTER TABLE grid_asset_presets DROP COLUMN IF EXISTS deleted_at;
