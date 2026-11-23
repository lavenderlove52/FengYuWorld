-- +goose Up
-- 权重改为 6 位小数：目标预算按金额反算时，2 位小数会让金额回弹几百上千元。
-- numeric(9,6) 表示最多 3 位整数 + 6 位小数，足够覆盖 0~100%（以及超配场景）。
ALTER TABLE grid_assets
  ALTER COLUMN weight TYPE NUMERIC(9,6);

-- +goose Down
ALTER TABLE grid_assets ALTER COLUMN weight TYPE NUMERIC(6,2);
