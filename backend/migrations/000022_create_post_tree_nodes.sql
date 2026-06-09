-- +goose Up
CREATE TABLE post_tree_nodes (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  parent_id BIGINT NULL REFERENCES post_tree_nodes(id) ON DELETE CASCADE,
  node_type VARCHAR(20) NOT NULL,
  name VARCHAR(120) NOT NULL DEFAULT '',
  post_id BIGINT NULL REFERENCES posts(id) ON DELETE CASCADE,
  sort_order INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT post_tree_nodes_type_check CHECK (node_type IN ('folder', 'post')),
  CONSTRAINT post_tree_nodes_shape_check CHECK (
    (node_type = 'folder' AND post_id IS NULL AND name <> '') OR
    (node_type = 'post' AND post_id IS NOT NULL)
  )
);

CREATE INDEX idx_post_tree_nodes_user_parent ON post_tree_nodes(user_id, parent_id, sort_order, id);
CREATE UNIQUE INDEX idx_post_tree_nodes_user_post ON post_tree_nodes(user_id, post_id) WHERE node_type = 'post';

-- +goose Down
DROP TABLE IF EXISTS post_tree_nodes;
