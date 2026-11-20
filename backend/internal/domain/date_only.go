package domain

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// DateOnly 是只含日期（yyyy-MM-dd）的类型。
// PostgreSQL 的 DATE 列不能直接扫进 *string，这里做一层转换；
// JSON 里仍然是 "2026-11-19" 这样的字符串。
type DateOnly struct {
	Time  time.Time
	Valid bool
}

const dateLayout = "2006-01-02"

func NewDateOnly(t time.Time) DateOnly {
	return DateOnly{Time: t, Valid: true}
}

// ParseDateOnly 解析 yyyy-MM-dd，失败返回零值。
func ParseDateOnly(s string) DateOnly {
	s = strings.TrimSpace(s)
	if s == "" {
		return DateOnly{}
	}
	if len(s) > 10 {
		s = s[:10]
	}
	t, err := time.ParseInLocation(dateLayout, s, time.Local)
	if err != nil {
		return DateOnly{}
	}
	return DateOnly{Time: t, Valid: true}
}

func (d DateOnly) String() string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format(dateLayout)
}

func (d DateOnly) MarshalJSON() ([]byte, error) {
	if !d.Valid {
		return []byte("null"), nil
	}
	return []byte(`"` + d.String() + `"`), nil
}

func (d *DateOnly) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		*d = DateOnly{}
		return nil
	}
	*d = ParseDateOnly(s)
	return nil
}

func (d DateOnly) Value() (driver.Value, error) {
	if !d.Valid {
		return nil, nil
	}
	return d.Time.Format(dateLayout), nil
}

func (d *DateOnly) Scan(value any) error {
	if value == nil {
		*d = DateOnly{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*d = DateOnly{Time: v, Valid: true}
		return nil
	case []byte:
		*d = ParseDateOnly(string(v))
		return nil
	case string:
		*d = ParseDateOnly(v)
		return nil
	}
	return fmt.Errorf("DateOnly: 无法转换 %T", value)
}
