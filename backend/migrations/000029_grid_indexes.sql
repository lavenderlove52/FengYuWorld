-- +goose Up
-- 执行记录按 (executed_at DESC, id DESC) 排序，索引缺 id DESC 时会多一次 incremental sort。
DROP INDEX IF EXISTS idx_grid_executions_plan;
CREATE INDEX idx_grid_executions_plan ON grid_plan_executions(plan_id, executed_at DESC, id DESC);

-- 列表页汇总只扫 invested / market_value 两列，做成覆盖索引可以免回表。
-- 排序用的是 (plan_id, sort_order, id)，这里保留同前缀并用 INCLUDE 补上汇总需要的列。
DROP INDEX IF EXISTS idx_grid_assets_plan;
CREATE INDEX idx_grid_assets_plan ON grid_assets(plan_id, sort_order, id) INCLUDE (invested, market_value);

-- +goose Down
DROP INDEX IF EXISTS idx_grid_assets_plan;
CREATE INDEX idx_grid_assets_plan ON grid_assets(plan_id, sort_order, id);
DROP INDEX IF EXISTS idx_grid_executions_plan;
CREATE INDEX idx_grid_executions_plan ON grid_plan_executions(plan_id, executed_at DESC);
