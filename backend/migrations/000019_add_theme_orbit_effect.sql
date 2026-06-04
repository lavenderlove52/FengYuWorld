-- +goose Up
UPDATE user_themes
SET effect_config = jsonb_set(
  effect_config,
  '{orbit}',
  '{"type":"off","color":"#6366f1","count":8,"size":5,"opacity":0.8,"speed":1,"radius":48}'::jsonb,
  true
)
WHERE NOT effect_config ? 'orbit';

ALTER TABLE user_themes
ALTER COLUMN effect_config SET DEFAULT '{
  "scope": "all",
  "background": {"type": "off", "color": "#6366f1", "count": 20, "size": 20, "opacity": 0.25, "speed": 0.5},
  "orbit": {"type": "off", "color": "#6366f1", "count": 8, "size": 5, "opacity": 0.8, "speed": 1, "radius": 48},
  "cursor": {"type": "off", "color": "#6366f1", "count": 16, "size": 5, "opacity": 0.7, "duration": 700},
  "click": {"type": "off", "color": "#ec4899", "count": 8, "size": 16, "opacity": 0.9, "duration": 900}
}'::jsonb;

-- +goose Down
ALTER TABLE user_themes
ALTER COLUMN effect_config SET DEFAULT '{
  "scope": "all",
  "background": {"type": "off", "color": "#6366f1", "count": 20, "size": 20, "opacity": 0.25, "speed": 0.5},
  "cursor": {"type": "off", "color": "#6366f1", "count": 16, "size": 5, "opacity": 0.7, "duration": 700},
  "click": {"type": "off", "color": "#ec4899", "count": 8, "size": 16, "opacity": 0.9, "duration": 900}
}'::jsonb;

UPDATE user_themes
SET effect_config = effect_config - 'orbit';
