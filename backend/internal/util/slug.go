package util

import (
	"regexp"
	"strings"
)

var (
	slugRe           = regexp.MustCompile(`[^a-z0-9\-]+`)
	slugAllNumericRe = regexp.MustCompile(`^\d+$`)
)

func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = slugRe.ReplaceAllString(s, "")
	return EnsurePostSlugBase(s)
}

// SlugIsAllNumeric 是否为非空的纯数字 slug（与文章 ID 路径冲突）
func SlugIsAllNumeric(s string) bool {
	s = strings.TrimSpace(s)
	return s != "" && slugAllNumericRe.MatchString(s)
}

// EnsurePostSlugBase 自动生成的 slug 若为纯数字则追加 -post
func EnsurePostSlugBase(base string) string {
	if SlugIsAllNumeric(base) {
		return base + "-post"
	}
	return base
}
