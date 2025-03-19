package sheetah

import (
	"fmt"
	"time"
)

func ParseTimeByString(str string, loc *time.Location) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05Z07:00", // RFC3339
		"2006-01-02 15:04:05Z07:00",
		"2006/01/02T15:04:05Z07:00",
		"2006/01/02 15:04:05Z07:00",
	}
	for _, format := range formats {
		parsed, err := time.Parse(format, str)
		if err == nil {
			return parsed, nil
		}
	}

	formats = []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006/01/02T15:04:05",
		"2006/01/02 15:04:05",
		"2006-01-02",
		"2006/01/02",
	}
	for _, format := range formats {
		parsed, err := time.ParseInLocation(format, str, loc)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("failed to parse time: %s", str)
}

func ParseTimeBySerialNumber(serial float64, loc *time.Location) time.Time {
	duration := time.Duration(serial * float64(24*60*60*time.Second))
	t := time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC).Add(duration)
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}
