package service

import (
	"time"

	"my-blog/backend/internal/domain"
)

const dateLayout = "2006-01-02"

var frequencyDays = map[string]int{
	"weekly":   7,
	"biweekly": 14,
	"monthly":  30,
}

// advanceByFrequency 按频率把定投日往后推，直到日期超过今天。
// 与前端 format.ts 的 advanceByFrequency 保持一致：只往后推、不往回拉。
func advanceByFrequency(base domain.DateOnly, frequency string) domain.DateOnly {
	days, ok := frequencyDays[frequency]
	if !ok {
		days = 30
	}

	start := base.Time
	if !base.Valid {
		start = time.Now()
	}

	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)

	next := start.AddDate(0, 0, days)
	for guard := 0; guard < 60 && !next.After(today); guard++ {
		next = next.AddDate(0, 0, days)
	}
	return domain.NewDateOnly(next)
}
