//go:build windows

package monitor

import (
	"context"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"netwatcher/next/internal/domain"
)

var windowsLatencyPattern = regexp.MustCompile(`(?i)(?:time|s.re)?\s*[=<]\s*(\d+(?:[\.,]\d+)?)\s*ms`)

func checkPing(ctx context.Context, target domain.Target, timeout time.Duration) domain.Result {
	started := time.Now()
	cmd := exec.CommandContext(ctx, "ping", "-n", "1", "-w", strconv.Itoa(int(timeout.Milliseconds())), target.Host)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	output, err := cmd.CombinedOutput()
	latency := float64(time.Since(started).Microseconds()) / 1000
	if match := windowsLatencyPattern.FindStringSubmatch(string(output)); len(match) == 2 {
		value := strings.ReplaceAll(match[1], ",", ".")
		if parsed, parseErr := strconv.ParseFloat(value, 64); parseErr == nil {
			latency = parsed
		}
	}
	return domain.Result{Timestamp: time.Now(), Target: target, Success: err == nil, Latency: latency, Message: messageFor(err, "Reply received")}
}
