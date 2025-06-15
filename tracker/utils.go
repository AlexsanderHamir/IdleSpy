package tracker

import (
	"runtime"
	"strconv"
	"strings"
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
