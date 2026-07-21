//go:build !windows

package monitor

import (
	"context"
	"os/exec"
	"strconv"
	"time"

	"netwatcher/next/internal/domain"
)

func checkPing(ctx context.Context, target domain.Target, timeout time.Duration) domain.Result {
	started := time.Now()
	seconds := int(timeout.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	err := exec.CommandContext(ctx, "ping", "-c", "1", "-W", strconv.Itoa(seconds), target.Host).Run()
	latency := float64(time.Since(started).Microseconds()) / 1000
	return domain.Result{Timestamp: time.Now(), Target: target, Success: err == nil, Latency: latency, Message: messageFor(err, "Reply received")}
}
