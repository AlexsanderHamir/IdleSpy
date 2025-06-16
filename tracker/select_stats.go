package tracker

import "time"

// GetCaseHits returns the number of times this case was hit
func (ss *SelectStats) GetCaseHits() int {
	return ss.CaseHits
}

// GetCaseTime returns the total time spent in this case
func (ss *SelectStats) GetCaseTime() time.Duration {
	return ss.BlockedCaseTime
}

// Get Average
func (ss *SelectStats) GetAverage() time.Duration {
	return ss.BlockedCaseTime / time.Duration(ss.CaseHits)
}
