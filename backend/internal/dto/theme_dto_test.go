package dto

import (
	"encoding/json"
	"testing"
)

func TestThemeEffectConfigRequestRejectsUnknownFields(t *testing.T) {
	tests := []string{
		`{
			"scope":"all",
			"background":{"type":"off","color":"#fff","count":1,"size":4,"opacity":1,"speed":1},
			"cursor":{"type":"off","color":"#fff","count":1,"size":1,"opacity":1,"duration":100},
			"click":{"type":"off","color":"#fff","count":1,"size":4,"opacity":1,"duration":200},
			"code":"alert(1)"
		}`,
		`{
			"scope":"all",
			"background":{"type":"off","color":"#fff","count":1,"size":4,"opacity":1,"speed":1,"javascript":"alert(1)"},
			"cursor":{"type":"off","color":"#fff","count":1,"size":1,"opacity":1,"duration":100},
			"click":{"type":"off","color":"#fff","count":1,"size":4,"opacity":1,"duration":200}
		}`,
	}

	for _, payload := range tests {
		var config ThemeEffectConfigRequest
		if err := json.Unmarshal([]byte(payload), &config); err == nil {
			t.Fatal("expected unknown field to be rejected")
		}
	}
}

func TestThemeHeroConfigRequestRejectsCodeFields(t *testing.T) {
	payloads := []string{
		`{
			"enabled":true,
			"image_ids":[1],
			"transition":"fade",
			"interval_ms":6000,
			"overlay_opacity":0.4,
			"hero_height_vh":100,
			"page_width_px":1280,
			"article_list_width_px":960,
			"article_layout":"list",
			"code":"alert(1)"
		}`,
		`{
			"enabled":true,
			"image_ids":[1],
			"transition":"fade",
			"interval_ms":6000,
			"overlay_opacity":0.4,
			"hero_height_vh":100,
			"page_width_px":1280,
			"article_list_width_px":960,
			"article_layout":"list",
			"javascript":"alert(1)"
		}`,
	}

	for _, payload := range payloads {
		var config ThemeHeroConfigRequest
		if err := json.Unmarshal([]byte(payload), &config); err == nil {
			t.Fatal("expected unknown code field to be rejected")
		}
	}
}
