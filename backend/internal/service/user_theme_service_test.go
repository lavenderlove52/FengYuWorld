package service

import (
	"errors"
	"testing"

	"my-blog/backend/internal/domain"
)

func TestValidateEffectConfig(t *testing.T) {
	t.Run("accepts trusted defaults", func(t *testing.T) {
		if err := validateEffectConfig(domain.DefaultThemeEffectConfig()); err != nil {
			t.Fatalf("validate defaults: %v", err)
		}
	})

	t.Run("accepts enabled effects within bounds", func(t *testing.T) {
		config := domain.DefaultThemeEffectConfig()
		config.Scope = "post"
		config.Background.Type = "magnetic-shapes"
		config.Orbit.Type = "particle-orbit"
		config.Cursor.Type = "hue-trail"
		config.Click.Type = "flower"

		if err := validateEffectConfig(config); err != nil {
			t.Fatalf("validate enabled effects: %v", err)
		}
	})

	t.Run("accepts legacy effects after orbit default is supplied", func(t *testing.T) {
		config := domain.DefaultThemeEffectConfig()
		config.Orbit = domain.ThemeOrbitEffect{}
		if config.Orbit.Type == "" {
			config.Orbit = domain.DefaultThemeEffectConfig().Orbit
		}
		if err := validateEffectConfig(config); err != nil {
			t.Fatalf("validate legacy effects: %v", err)
		}
	})

	tests := []struct {
		name   string
		change func(*domain.ThemeEffectConfig)
	}{
		{"unknown scope", func(c *domain.ThemeEffectConfig) { c.Scope = "admin" }},
		{"unknown background", func(c *domain.ThemeEffectConfig) { c.Background.Type = "script" }},
		{"unknown orbit", func(c *domain.ThemeEffectConfig) { c.Orbit.Type = "script" }},
		{"invalid orbit radius", func(c *domain.ThemeEffectConfig) { c.Orbit.Radius = 500 }},
		{"invalid color", func(c *domain.ThemeEffectConfig) {
			c.Click.Type = "heart"
			c.Click.Color = "javascript:alert(1)"
		}},
		{"too many particles", func(c *domain.ThemeEffectConfig) {
			c.Cursor.Type = "particle-trail"
			c.Cursor.Count = 1000
		}},
		{"invalid duration", func(c *domain.ThemeEffectConfig) {
			c.Click.Type = "dot"
			c.Click.Duration = 10
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := domain.DefaultThemeEffectConfig()
			test.change(&config)
			if err := validateEffectConfig(config); !errors.Is(err, ErrInvalidEffectConfig) {
				t.Fatalf("expected ErrInvalidEffectConfig, got %v", err)
			}
		})
	}
}

func TestValidateHeroConfig(t *testing.T) {
	t.Run("accepts trusted defaults", func(t *testing.T) {
		if err := validateHeroConfig(domain.DefaultThemeHeroConfig()); err != nil {
			t.Fatalf("validate defaults: %v", err)
		}
	})

	t.Run("accepts enabled hero", func(t *testing.T) {
		config := domain.DefaultThemeHeroConfig()
		config.Enabled = true
		config.ImageIDs = []int64{4, 2}
		config.Transition = "slices"
		config.ArticleLayout = "grid"
		config.PageWidthPX = 720
		config.ArticleListWidthPX = 1400
		if err := validateHeroConfig(config); err != nil {
			t.Fatalf("validate enabled hero: %v", err)
		}
	})

	t.Run("accepts interlaced transition", func(t *testing.T) {
		config := domain.DefaultThemeHeroConfig()
		config.Transition = "interlaced"
		if err := validateHeroConfig(config); err != nil {
			t.Fatalf("validate interlaced transition: %v", err)
		}
	})

	tests := []struct {
		name   string
		change func(*domain.ThemeHeroConfig)
	}{
		{"enabled without image", func(c *domain.ThemeHeroConfig) { c.Enabled = true }},
		{"more than five images", func(c *domain.ThemeHeroConfig) { c.ImageIDs = []int64{1, 2, 3, 4, 5, 6} }},
		{"duplicate image", func(c *domain.ThemeHeroConfig) { c.ImageIDs = []int64{1, 1} }},
		{"invalid image", func(c *domain.ThemeHeroConfig) { c.ImageIDs = []int64{-1} }},
		{"unknown transition", func(c *domain.ThemeHeroConfig) { c.Transition = "javascript" }},
		{"short interval", func(c *domain.ThemeHeroConfig) { c.IntervalMS = 100 }},
		{"short slice interval", func(c *domain.ThemeHeroConfig) {
			c.Transition = "slices"
			c.IntervalMS = 3000
		}},
		{"invalid overlay", func(c *domain.ThemeHeroConfig) { c.OverlayOpacity = 1 }},
		{"invalid hero height", func(c *domain.ThemeHeroConfig) { c.HeroHeightVH = 20 }},
		{"invalid page width", func(c *domain.ThemeHeroConfig) { c.PageWidthPX = 500 }},
		{"invalid article width", func(c *domain.ThemeHeroConfig) { c.ArticleListWidthPX = 1800 }},
		{"unknown article layout", func(c *domain.ThemeHeroConfig) { c.ArticleLayout = "script" }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := domain.DefaultThemeHeroConfig()
			test.change(&config)
			if err := validateHeroConfig(config); !errors.Is(err, ErrInvalidHeroConfig) {
				t.Fatalf("expected ErrInvalidHeroConfig, got %v", err)
			}
		})
	}
}
