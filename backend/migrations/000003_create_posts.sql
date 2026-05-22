-- +goose Up
CREATE TABLE posts (
  id BIGSERIAL PRIMARY KEY,
  author_id BIGINT NOT NULL REFERENCES users(id),
  title VARCHAR(200) NOT NULL,
  slug VARCHAR(220) NOT NULL DEFAULT '',
  cover_image_url VARCHAR(500) NOT NULL DEFAULT '',
  content_md TEXT NOT NULL,
  content_html TEXT NOT NULL,
  category_id BIGINT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'draft',
  visibility VARCHAR(20) NOT NULL DEFAULT 'private',
  allow_comment BOOLEAN NOT NULL DEFAULT true,
  view_count BIGINT NOT NULL DEFAULT 0,
  is_deleted BOOLEAN NOT NULL DEFAULT false,
  published_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT posts_status_check CHECK (status IN ('draft', 'published')),
  CONSTRAINT posts_visibility_check CHECK (visibility IN ('public', 'private', 'friends', 'specified'))
);

CREATE INDEX idx_posts_author_status_deleted ON posts(author_id, status, is_deleted);
CREATE INDEX idx_posts_public_list ON posts(status, visibility, is_deleted, published_at DESC);
CREATE INDEX idx_posts_category_id ON posts(category_id);

CREATE TABLE post_visible_users (
  post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (post_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS post_visible_users;
DROP TABLE IF EXISTS posts;
