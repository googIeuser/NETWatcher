package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// readGraphHistory restores the last selected graph window from the local CSV
// logs. This makes 30-minute, 1-hour and 24-hour ranges show persisted data
// from previous app sessions instead of only stretching the current session.
func readGraphHistory(logDir string, since time.Time) (map[string][]Sample, error) {
	paths, err := filepath.Glob(filepath.Join(logDir, "samples_*.csv"))
	if err != nil {
		return nil, err
	}
	history := make(map[string][]Sample)
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		reader := csv.NewReader(file)
		reader.Comma = ';'
		reader.FieldsPerRecord = -1
		rows, readErr := reader.ReadAll()
		_ = file.Close()
		if readErr != nil {
			continue
		}
		for index, row := range rows {
			if index == 0 || len(row) < 7 {
				continue
			}
			row[0] = strings.TrimPrefix(row[0], "\ufeff")
			timestamp, err := time.Parse(time.RFC3339, row[0])
			if err != nil || timestamp.Before(since) {
				continue
			}
			host := strings.TrimSpace(row[2])
			if host == "" {
				continue
			}
			success, _ := strconv.ParseBool(strings.TrimSpace(row[4]))
			latency, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(row[5]), ",", "."), 64)
			history[host] = append(history[host], Sample{Time: timestamp, Latency: latency, Success: success})
		}
	}
	for host, samples := range history {
		sort.Slice(samples, func(i, j int) bool { return samples[i].Time.Before(samples[j].Time) })
		if len(samples) > maxHistory {
			samples = samples[len(samples)-maxHistory:]
		}
		history[host] = samples
	}
	return history, nil
}

func graphRangeLabel(minutes int) string {
	switch normalizeGraphRange(minutes) {
	case 30:
		return "Last 30 minutes"
	case 60:
		return "Last 1 hour"
	case 1440:
		return "Last 24 hours"
	default:
		return "Last 5 minutes"
	}
}

func formatGraphAxisTime(value time.Time, minutes int) string {
	if normalizeGraphRange(minutes) == 1440 {
		return value.Format("Jan 02 15:04")
	}
	return value.Format("15:04:05")
}
