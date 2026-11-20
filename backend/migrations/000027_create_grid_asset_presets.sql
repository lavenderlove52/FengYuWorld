-- +goose Up
-- 用户自建的标的（名字/代码手输或从计划里保存下来的），只用于下拉联想，不参与计划计算。
CREATE TABLE grid_asset_presets (
  id         BIGSERIAL   PRIMARY KEY,
  user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  market     VARCHAR(20) NOT NULL DEFAULT 'cn',
  name       VARCHAR(80) NOT NULL,
  code       VARCHAR(60) NOT NULL DEFAULT '',
  leveraged  BOOLEAN     NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT grid_asset_presets_market_check CHECK (market IN ('cn', 'us')),
  CONSTRAINT grid_asset_presets_unique UNIQUE (user_id, market, name)
);

-- UNIQUE (user_id, market, name) 已提供同前缀索引，无需再建一份

-- +goose Down
DROP TABLE IF EXISTS grid_asset_presets;
