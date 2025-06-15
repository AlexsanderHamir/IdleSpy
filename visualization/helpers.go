package visualization

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

func ConvertToSeconds(value float64, unit string) float64 {
	switch unit {
	case "s", "sec", "second", "seconds":
		return value
	case "ms", "millisecond", "milliseconds":
		return value / 1_000
	case "µs", "us", "microsecond", "microseconds":
		return value / 1_000_000
	case "ns", "nanosecond", "nanoseconds":
		return value / 1_000_000_000
	default:
		return value
	}
}

var (
	goroutinePattern   = regexp.MustCompile(`Goroutine (\d+):`)
	lifetimePattern    = regexp.MustCompile(`Lifetime:\s*([\d.]+)([a-zµ]+)`)
	blockedTimePattern = regexp.MustCompile(`Total Select Blocked Time:\s*([\d.]+)([a-zµ]+)`)
)

func isGoroutineLine(line string) bool {
	return goroutinePattern.MatchString(line)
}

func extractGoroutineID(line string) (int, error) {
	matches := goroutinePattern.FindStringSubmatch(line)
	if matches == nil {
		return 0, fmt.Errorf("invalid goroutine line: %s", line)
	}
	return strconv.Atoi(matches[1])
}

func isLifetimeLine(line string) bool {
	return lifetimePattern.MatchString(line)
}

func isBlockedTimeLine(line string) bool {
	return blockedTimePattern.MatchString(line)
}

func extractDuration(line, pattern string) (time.Duration, error) {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 3 {
		return 0, fmt.Errorf("invalid duration format: %s", line)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, err
	}
	unit := matches[2]
	seconds := ConvertToSeconds(value, unit)
	return time.Duration(seconds * float64(time.Second)), nil
}
