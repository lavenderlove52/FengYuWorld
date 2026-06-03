-- +goose Up
UPDATE posts SET show_toc = true WHERE show_toc = false;
ALTER TABLE posts ALTER COLUMN show_toc SET DEFAULT true;

-- +goose Down
ALTER TABLE posts ALTER COLUMN show_toc SET DEFAULT false;
