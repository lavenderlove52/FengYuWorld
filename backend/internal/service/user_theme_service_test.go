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
		config.Background.Type = "floating-shapes"
		config.Cursor.Type = "particle-trail"
		config.Click.Type = "star"

		if err := validateEffectConfig(config); err != nil {
			t.Fatalf("validate enabled effects: %v", err)
		}
	})

	tests := []struct {
		name   string
		change func(*domain.ThemeEffectConfig)
	}{
		{"unknown scope", func(c *domain.ThemeEffectConfig) { c.Scope = "admin" }},
		{"unknown background", func(c *domain.ThemeEffectConfig) { c.Background.Type = "script" }},
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
