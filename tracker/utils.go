package tracker

import (
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/AlexsanderHamir/IdleSpy/sharedtypes"
)

// getGoroutineID returns a unique ID for the current goroutine
func getGoroutineID() GoroutineId {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.ParseInt(idField, 10, 64)
	if err != nil {
		panic("cannot get goroutine id: " + err.Error())
	}
	return GoroutineId(id)
}

func (gm *GoroutineManager) handleTextActions() {
	allStats := gm.GetAllStats()
	switch gm.Action {
	case PrintAndSave:
		PrintAndSaveStatsText(allStats, ".visualization")
		SaveStatsJSON(allStats, ".internal")
	case Save:
		SaveStatsText(allStats, ".visualization")
		SaveStatsJSON(allStats, ".internal")
	case Print:
		PrintStatsText(allStats, ".visualization")
	}
}

func (gm *GoroutineManager) handleJsonActions() {
	allStats := gm.GetAllStats()
	switch gm.Action {
	case PrintAndSave:
		PrintAndSaveStatsJSON(allStats, ".internal")
	case Save:
		SaveStatsJSON(allStats, ".internal")
	case Print:
		PrintStatsJSON(allStats, ".internal")
	}
}

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

// AggregateCaseStats combines statistics for cases with the same name
func AggregateCaseStats(caseStats []*sharedtypes.CaseJSON) map[string]*sharedtypes.CaseJSON {
	aggregatedStats := make(map[string]*sharedtypes.CaseJSON)
	for _, stat := range caseStats {
		if existing, exists := aggregatedStats[stat.CaseName]; exists {
			existing.Hits += stat.Hits
			existing.TotalBlockedTime += stat.TotalBlockedTime
			existing.AvgBlockedTime += stat.AvgBlockedTime

			if stat.Percentile90 > existing.Percentile90 {
				existing.Percentile90 = stat.Percentile90
			}
			if stat.Percentile99 > existing.Percentile99 {
				existing.Percentile99 = stat.Percentile99
			}
		} else {
			aggregatedStats[stat.CaseName] = &sharedtypes.CaseJSON{
				CaseName:         stat.CaseName,
				Hits:             stat.Hits,
				TotalBlockedTime: stat.TotalBlockedTime,
				AvgBlockedTime:   stat.AvgBlockedTime,
				Percentile90:     stat.Percentile90,
				Percentile99:     stat.Percentile99,
			}
		}
	}
	return aggregatedStats
}

// SortCaseStats sorts the aggregated statistics based on the visualization type
func SortCaseStats(stats []*sharedtypes.CaseJSON, visType sharedtypes.VisualizationType) {
	sort.Slice(stats, func(i, j int) bool {
		switch visType {
		case sharedtypes.TotalBlockedTime:
			return stats[i].TotalBlockedTime > stats[j].TotalBlockedTime
		case sharedtypes.AverageTime:
			return stats[i].AvgBlockedTime > stats[j].AvgBlockedTime
		case sharedtypes.Percentile90:
			return stats[i].Percentile90 > stats[j].Percentile90
		case sharedtypes.Percentile99:
			return stats[i].Percentile99 > stats[j].Percentile99
		case sharedtypes.TotalHits:
			return stats[i].Hits > stats[j].Hits
		default:
			return false
		}
	})
}

// GetMaxValue returns the maximum value for the given visualization type
func GetMaxValue(stats []*sharedtypes.CaseJSON, visType sharedtypes.VisualizationType) float64 {
	if len(stats) == 0 {
		return 0
	}
	switch visType {
	case sharedtypes.TotalBlockedTime:
		return float64(stats[0].TotalBlockedTime)
	case sharedtypes.AverageTime:
		return float64(stats[0].AvgBlockedTime)
	case sharedtypes.Percentile90:
		return float64(stats[0].Percentile90)
	case sharedtypes.Percentile99:
		return float64(stats[0].Percentile99)
	case sharedtypes.TotalHits:
		return float64(stats[0].Hits)
	default:
		return 0
	}
}

// GetValueForCase returns the value for a case based on the visualization type
func GetValueForCase(stat *sharedtypes.CaseJSON, visType sharedtypes.VisualizationType) float64 {
	switch visType {
	case sharedtypes.TotalBlockedTime:
		return float64(stat.TotalBlockedTime)
	case sharedtypes.AverageTime:
		return float64(stat.AvgBlockedTime)
	case sharedtypes.Percentile90:
		return float64(stat.Percentile90)
	case sharedtypes.Percentile99:
		return float64(stat.Percentile99)
	case sharedtypes.TotalHits:
		return float64(stat.Hits)
	default:
		return 0
	}
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}
