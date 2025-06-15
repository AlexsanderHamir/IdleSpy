package sharedtypes

// VisualizationType represents the type of visualization to use
type VisualizationType int

const (
	TotalBlockedTime VisualizationType = iota
	AverageTime
	Percentile90
	Percentile99
	TotalHits
)

func (vt VisualizationType) String() string {
	switch vt {
	case TotalBlockedTime:
		return "Total Blocked Time"
	case AverageTime:
		return "Average Blocked Time"
	case Percentile90:
		return "90th Percentile"
	case Percentile99:
		return "99th Percentile"
	case TotalHits:
		return "Total Hits"
	default:
		return "Unknown"
	}
}

// CaseJSON represents statistics for a single select case in JSON format
type CaseJSON struct {
	CaseName         string `json:"case_name"`
	Hits             int64  `json:"hits"`
	TotalBlockedTime int64  `json:"total_blocked_time"`
	AvgBlockedTime   int64  `json:"average_blocked_time"`
	Percentile90     int64  `json:"percentile_90"`
	Percentile99     int64  `json:"percentile_99"`
}
