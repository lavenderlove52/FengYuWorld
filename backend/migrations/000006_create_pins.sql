-- +goose Up
CREATE TABLE site_pins (
  id           BIGSERIAL PRIMARY KEY,
  post_id      BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  display_order INT NOT NULL DEFAULT 0,
  pinned_by    BIGINT NOT NULL REFERENCES users(id),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uq_site_pins_post UNIQUE (post_id)
);

CREATE TABLE user_pins (
  id           BIGSERIAL PRIMARY KEY,
  user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  post_id      BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  display_order INT NOT NULL DEFAULT 0,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uq_user_pins_user_post UNIQUE (user_id, post_id)
);

ALTER TABLE users ADD COLUMN can_site_pin BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS can_site_pin;
DROP TABLE IF EXISTS user_pins;
DROP TABLE IF EXISTS site_pins;
