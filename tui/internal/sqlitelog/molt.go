package sqlitelog

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// QueryMoltSessionWindows fetches the latest two psyche_molt timestamps from
// the sqlite sidecar. It returns the current session lower bound, the previous
// session lower bound, and the previous session upper bound.
func QueryMoltSessionWindows(agentDir string) (currentSince, lastSince, lastBefore time.Time, ok bool, err error) {
	db := DBPath(agentDir)
	if _, err := os.Stat(db); err != nil {
		return time.Time{}, time.Time{}, time.Time{}, false, fmt.Errorf("sqlite sidecar not found: %s", db)
	}
	bin, err := findSQLite3()
	if err != nil {
		return time.Time{}, time.Time{}, time.Time{}, false, err
	}

	const sql = `SELECT ts FROM events WHERE type='psyche_molt' ORDER BY ts DESC LIMIT 2`
	out, err := exec.Command(bin, "-separator", "\x1f", db, sql).Output()
	if err != nil {
		msg := ""
		if ee, ok := err.(*exec.ExitError); ok {
			msg = strings.TrimSpace(string(ee.Stderr))
		}
		if msg != "" {
			return time.Time{}, time.Time{}, time.Time{}, false, fmt.Errorf("sqlite3: %s", msg)
		}
		return time.Time{}, time.Time{}, time.Time{}, false, fmt.Errorf("sqlite3 query failed: %w", err)
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return time.Time{}, time.Time{}, time.Time{}, true, nil
	}
	lines := strings.Split(raw, "\n")
	latest, err := strconv.ParseFloat(strings.TrimSpace(lines[0]), 64)
	if err != nil || latest <= 0 {
		return time.Time{}, time.Time{}, time.Time{}, false, fmt.Errorf("invalid psyche_molt ts %q", strings.TrimSpace(lines[0]))
	}
	currentSince = unixFloatTimeUTC(latest)
	if len(lines) > 1 {
		previous, err := strconv.ParseFloat(strings.TrimSpace(lines[1]), 64)
		if err != nil || previous <= 0 {
			return time.Time{}, time.Time{}, time.Time{}, false, fmt.Errorf("invalid previous psyche_molt ts %q", strings.TrimSpace(lines[1]))
		}
		lastSince = unixFloatTimeUTC(previous)
		lastBefore = currentSince
	}
	return currentSince, lastSince, lastBefore, true, nil
}

func unixFloatTimeUTC(ts float64) time.Time {
	sec := int64(ts)
	nsec := int64((ts - float64(sec)) * 1e9)
	return time.Unix(sec, nsec).UTC()
}
