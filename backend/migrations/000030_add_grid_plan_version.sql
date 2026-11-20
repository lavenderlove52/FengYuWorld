-- +goose Up
-- 乐观锁版本号：保存时带上读到的版本，写库时比对，避免旧标签页覆盖新数据
ALTER TABLE grid_plans ADD COLUMN version BIGINT NOT NULL DEFAULT 1;

-- +goose Down
ALTER TABLE grid_plans DROP COLUMN IF EXISTS version;
