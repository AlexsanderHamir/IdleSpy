package visualization

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
