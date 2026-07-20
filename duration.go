package main

import (
	"fmt"
	"time"
)

func formatDuration(d time.Duration, lang string) string {
	seconds := int(d.Seconds())
	if seconds < 0 {
		seconds = 0
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	if lang == "en" {
		if h > 0 {
			return fmt.Sprintf("%d h %d min %d sec", h, m, s)
		}
		if m > 0 {
			return fmt.Sprintf("%d min %d sec", m, s)
		}
		return fmt.Sprintf("%d sec", s)
	}
	if h > 0 {
		return fmt.Sprintf("%d sa %d dk %d sn", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%d dk %d sn", m, s)
	}
	return fmt.Sprintf("%d sn", s)
}
