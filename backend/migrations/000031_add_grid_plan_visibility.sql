-- +goose Up
-- 投资计划对外可见范围：默认私密，只有作者手动改才会出现在个人主页
ALTER TABLE grid_plans
  ADD COLUMN visibility VARCHAR(20) NOT NULL DEFAULT 'private';

ALTER TABLE grid_plans
  ADD CONSTRAINT grid_plans_visibility_check
  CHECK (visibility IN ('private', 'login', 'friends', 'specified', 'public'));

-- 指定可见的好友，和 post_visible_users 同构
CREATE TABLE grid_plan_visible_users (
  plan_id    BIGINT      NOT NULL REFERENCES grid_plans(id) ON DELETE CASCADE,
  user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (plan_id, user_id)
);

-- 对外展示：作者可见，或已登录且计划登录可见 / 好友可见 / 指定可见
CREATE INDEX idx_grid_plans_visibility ON grid_plans(user_id, visibility);

-- 个人主页是否展示「投资计划」页签（默认关闭，不占位置）
ALTER TABLE user_themes
  ADD COLUMN show_plans_public BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE user_themes DROP COLUMN IF EXISTS show_plans_public;
DROP TABLE IF EXISTS grid_plan_visible_users;
DROP INDEX IF EXISTS idx_grid_plans_visibility;
ALTER TABLE grid_plans DROP CONSTRAINT IF EXISTS grid_plans_visibility_check;
ALTER TABLE grid_plans DROP COLUMN IF EXISTS visibility;
