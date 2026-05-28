-- +goose Up
CREATE TABLE user_themes (
  user_id          BIGINT       PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  preset           VARCHAR(50)  NOT NULL DEFAULT 'default',
  accent_color     VARCHAR(20)  NOT NULL DEFAULT '',
  font_family      VARCHAR(100) NOT NULL DEFAULT '',
  banner_url       VARCHAR(500) NOT NULL DEFAULT '',
  layout           VARCHAR(20)  NOT NULL DEFAULT 'grid',
  show_pins_public BOOLEAN      NOT NULL DEFAULT false,
  custom_css       TEXT         NOT NULL DEFAULT '',
  updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS user_themes;
