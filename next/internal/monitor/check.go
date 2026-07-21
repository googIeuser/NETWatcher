package monitor

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"netwatcher/next/internal/domain"
)

func checkTarget(ctx context.Context, target domain.Target, timeout time.Duration) domain.Result {
	switch target.Mode {
	case "tcp":
		return checkTCP(ctx, target, timeout)
	case "http", "https":
		return checkHTTP(ctx, target, timeout)
	default:
		return checkPing(ctx, target, timeout)
	}
}

func checkTCP(ctx context.Context, target domain.Target, timeout time.Duration) domain.Result {
	address := strings.TrimSpace(target.Host)
	if _, _, err := net.SplitHostPort(address); err != nil {
		address = net.JoinHostPort(address, "443")
	}
	started := time.Now()
	conn, err := (&net.Dialer{Timeout: timeout}).DialContext(ctx, "tcp", address)
	latency := float64(time.Since(started).Microseconds()) / 1000
	if conn != nil {
		_ = conn.Close()
	}
	return domain.Result{Timestamp: time.Now(), Target: target, Success: err == nil, Latency: latency, Message: messageFor(err, "TCP connection established")}
}

func checkHTTP(ctx context.Context, target domain.Target, timeout time.Duration) domain.Result {
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         (&net.Dialer{Timeout: timeout}).DialContext,
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
		TLSHandshakeTimeout: timeout,
		DisableKeepAlives:   true,
	}
	client := &http.Client{Transport: transport, Timeout: timeout}
	started := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, target.Host, nil)
	if err != nil {
		return domain.Result{Timestamp: time.Now(), Target: target, Message: err.Error()}
	}
	req.Header.Set("User-Agent", "NetWatcher/next")
	resp, err := client.Do(req)
	latency := float64(time.Since(started).Microseconds()) / 1000
	success := err == nil && resp != nil && resp.StatusCode < 500
	message := messageFor(err, "HTTP response received")
	if resp != nil {
		message = fmt.Sprintf("HTTP %d", resp.StatusCode)
		_ = resp.Body.Close()
	}
	return domain.Result{Timestamp: time.Now(), Target: target, Success: success, Latency: latency, Message: message}
}

func messageFor(err error, success string) string {
	if err != nil {
		return err.Error()
	}
	return success
}
