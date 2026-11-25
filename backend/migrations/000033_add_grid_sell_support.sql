-- +goose Up
-- 卖出支持：卖出只记「到账金额」，不推断卖的是哪一层，也不做分层成本。
-- realized_out 是累计卖出到账，invested 仍然是累计买入、只增不减。
ALTER TABLE grid_assets
  ADD COLUMN realized_out NUMERIC(18,2) NOT NULL DEFAULT 0;

ALTER TABLE grid_assets
  ADD CONSTRAINT grid_assets_realized_out_check CHECK (realized_out >= 0);

-- side 区分买入/卖出；point 记录成交时的点位快照，用来画买卖点图。
-- 老记录默认 buy、point 为空，图上不画点。
ALTER TABLE grid_plan_executions
  ADD COLUMN side VARCHAR(8) NOT NULL DEFAULT 'buy';

ALTER TABLE grid_plan_executions
  ADD COLUMN point NUMERIC(18,4) NULL;

ALTER TABLE grid_plan_executions
  ADD CONSTRAINT grid_plan_executions_side_check CHECK (side IN ('buy', 'sell'));

-- +goose Down
ALTER TABLE grid_plan_executions DROP CONSTRAINT IF EXISTS grid_plan_executions_side_check;
ALTER TABLE grid_plan_executions DROP COLUMN IF EXISTS point;
ALTER TABLE grid_plan_executions DROP COLUMN IF EXISTS side;
ALTER TABLE grid_assets DROP CONSTRAINT IF EXISTS grid_assets_realized_out_check;
ALTER TABLE grid_assets DROP COLUMN IF EXISTS realized_out;
