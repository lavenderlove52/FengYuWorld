package util

import (
	"regexp"
	"strings"
)

var slugRe = regexp.MustCompile(`[^a-z0-9\-]+`)

func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = slugRe.ReplaceAllString(s, "")
	return s
}
