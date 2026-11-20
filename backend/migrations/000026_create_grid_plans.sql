-- +goose Up
CREATE TABLE grid_plans (
  id                   BIGSERIAL      PRIMARY KEY,
  user_id              BIGINT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name                 VARCHAR(80)    NOT NULL,
  market               VARCHAR(20)    NOT NULL DEFAULT 'cn',
  currency             VARCHAR(10)    NOT NULL DEFAULT '¥',
  currency_name        VARCHAR(20)    NOT NULL DEFAULT '人民币',
  unit                 NUMERIC(18,4)  NOT NULL DEFAULT 100,
  total_budget         NUMERIC(18,2)  NOT NULL DEFAULT 0,
  frequency            VARCHAR(20)    NOT NULL DEFAULT 'monthly',
  next_date            DATE           NULL,
  gate_threshold       NUMERIC(6,2)   NOT NULL DEFAULT 20,
  gate_mode            VARCHAR(20)    NOT NULL DEFAULT 'grid-only',
  weight_method        VARCHAR(20)    NOT NULL DEFAULT 'target-weight',
  grid_aggregation     VARCHAR(20)    NOT NULL DEFAULT 'highest',
  unallocated_invested NUMERIC(18,2)  NOT NULL DEFAULT 0,
  sort_order           INT            NOT NULL DEFAULT 0,
  created_at           TIMESTAMPTZ    NOT NULL DEFAULT now(),
  updated_at           TIMESTAMPTZ    NOT NULL DEFAULT now(),
  CONSTRAINT grid_plans_frequency_check CHECK (frequency IN ('weekly', 'biweekly', 'monthly')),
  CONSTRAINT grid_plans_gate_mode_check CHECK (gate_mode IN ('alert', 'grid-only', 'block-all')),
  CONSTRAINT grid_plans_weight_method_check CHECK (weight_method IN ('target-weight', 'invested-weight')),
  CONSTRAINT grid_plans_aggregation_check CHECK (grid_aggregation IN ('highest', 'cumulative')),
  CONSTRAINT grid_plans_amounts_check CHECK (total_budget >= 0 AND unit > 0 AND unallocated_invested >= 0)
);

CREATE INDEX idx_grid_plans_user ON grid_plans(user_id, sort_order, id);

CREATE TABLE grid_assets (
  id              BIGSERIAL      PRIMARY KEY,
  plan_id         BIGINT         NOT NULL REFERENCES grid_plans(id) ON DELETE CASCADE,
  name            VARCHAR(80)    NOT NULL,
  code            VARCHAR(60)    NOT NULL DEFAULT '',
  leveraged       BOOLEAN        NOT NULL DEFAULT false,
  enabled         BOOLEAN        NOT NULL DEFAULT true,
  current_point   NUMERIC(18,4)  NULL,
  high_point      NUMERIC(18,4)  NULL,
  weight          NUMERIC(6,2)   NOT NULL DEFAULT 0,
  invested        NUMERIC(18,2)  NOT NULL DEFAULT 0,
  market_value    NUMERIC(18,2)  NOT NULL DEFAULT 0,
  base_amount     NUMERIC(18,2)  NOT NULL DEFAULT 0,
  grid            JSONB          NOT NULL DEFAULT '[]',
  executed_layers JSONB          NOT NULL DEFAULT '[]',
  sort_order      INT            NOT NULL DEFAULT 0,
  created_at      TIMESTAMPTZ    NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ    NOT NULL DEFAULT now(),
  CONSTRAINT grid_assets_amounts_check CHECK (weight >= 0 AND invested >= 0 AND market_value >= 0 AND base_amount >= 0),
  CONSTRAINT grid_assets_points_check CHECK (
    (current_point IS NULL OR current_point >= 0) AND (high_point IS NULL OR high_point >= 0)
  )
);

CREATE INDEX idx_grid_assets_plan ON grid_assets(plan_id, sort_order, id);

CREATE TABLE grid_plan_snapshots (
  id            BIGSERIAL      PRIMARY KEY,
  plan_id       BIGINT         NOT NULL REFERENCES grid_plans(id) ON DELETE CASCADE,
  snapshot_date DATE           NOT NULL,
  invested      NUMERIC(18,2)  NOT NULL,
  market_value  NUMERIC(18,2)  NOT NULL,
  pnl           NUMERIC(18,2)  NOT NULL,
  created_at    TIMESTAMPTZ    NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_grid_snapshots_plan_date ON grid_plan_snapshots(plan_id, snapshot_date);

CREATE TABLE grid_plan_executions (
  id             BIGSERIAL      PRIMARY KEY,
  plan_id        BIGINT         NOT NULL REFERENCES grid_plans(id) ON DELETE CASCADE,
  asset_id       BIGINT         NULL REFERENCES grid_assets(id) ON DELETE SET NULL,
  asset_name     VARCHAR(80)    NOT NULL,
  amount         NUMERIC(18,2)  NOT NULL,
  layer_keys     JSONB          NOT NULL DEFAULT '[]',
  was_regular    BOOLEAN        NOT NULL DEFAULT false,
  prev_next_date DATE           NULL,
  note           VARCHAR(200)   NOT NULL DEFAULT '',
  executed_at    TIMESTAMPTZ    NOT NULL DEFAULT now(),
  CONSTRAINT grid_plan_executions_amount_check CHECK (amount > 0)
);

CREATE INDEX idx_grid_executions_plan ON grid_plan_executions(plan_id, executed_at DESC);

-- +goose Down
DROP TABLE IF EXISTS grid_plan_executions;
DROP TABLE IF EXISTS grid_plan_snapshots;
DROP TABLE IF EXISTS grid_assets;
DROP TABLE IF EXISTS grid_plans;
