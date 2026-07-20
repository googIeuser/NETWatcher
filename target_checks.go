//go:build windows

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	checkModePing  = "ping"
	checkModeTCP   = "tcp"
	checkModeHTTP  = "http"
	checkModeHTTPS = "https"
)

func parseTargetSpec(raw string) Target {
	raw = strings.TrimSpace(raw)
	target := Target{Name: "Custom: " + raw, Host: raw, Kind: "internet", Mode: checkModePing, Custom: true}
	if raw == "" {
		return target
	}

	lower := strings.ToLower(raw)
	switch {
	case strings.HasPrefix(lower, "tcp://"):
		address := strings.TrimSpace(raw[len("tcp://"):])
		target.Mode = checkModeTCP
		target.Host = address
		target.Name = "TCP: " + address
	case strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://"):
		parsed, err := url.Parse(raw)
		if err == nil && parsed.Host != "" {
			target.Mode = strings.ToLower(parsed.Scheme)
			target.Host = raw
			target.Name = strings.ToUpper(parsed.Scheme) + ": " + parsed.Host
		}
	}
	return target
}

func targetConfigValue(t Target) string {
	switch t.Mode {
	case checkModeTCP:
		return "tcp://" + t.Host
	case checkModeHTTP, checkModeHTTPS:
		return t.Host
	default:
		return t.Host
	}
}

func checkTarget(t Target, timeout int, lang string) PingResult {
	switch t.Mode {
	case checkModeTCP:
		return checkTCP(t, timeout, lang)
	case checkModeHTTP, checkModeHTTPS:
		return checkHTTP(t, timeout, lang)
	default:
		return pingTarget(t, timeout, lang)
	}
}

func checkTCP(t Target, timeout int, lang string) PingResult {
	now := time.Now()
	if timeout < 200 {
		timeout = 200
	}
	address := strings.TrimSpace(t.Host)
	if _, _, err := net.SplitHostPort(address); err != nil {
		address = net.JoinHostPort(address, "443")
	}
	started := time.Now()
	conn, err := net.DialTimeout("tcp", address, time.Duration(timeout)*time.Millisecond)
	latency := float64(time.Since(started).Microseconds()) / 1000
	success := err == nil
	if conn != nil {
		_ = conn.Close()
	}
	message := tr(lang, "response_ok")
	if !success {
		message = "TCP connection failed"
	}
	return PingResult{Timestamp: now, Target: t, Success: success, Latency: latency, Message: message}
}

func checkHTTP(t Target, timeout int, lang string) PingResult {
	now := time.Now()
	if timeout < 200 {
		timeout = 200
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         (&net.Dialer{Timeout: time.Duration(timeout) * time.Millisecond}).DialContext,
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
		TLSHandshakeTimeout: time.Duration(timeout) * time.Millisecond,
		DisableKeepAlives:   true,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeout) * time.Millisecond,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	started := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, t.Host, nil)
	if err != nil {
		return PingResult{Timestamp: now, Target: t, Success: false, Message: err.Error()}
	}
	req.Header.Set("User-Agent", "NetWatcher/"+appVersion)
	resp, err := client.Do(req)
	latency := float64(time.Since(started).Microseconds()) / 1000
	success := err == nil && resp != nil && resp.StatusCode < 500
	message := tr(lang, "response_ok")
	if resp != nil {
		message = "HTTP " + strconv.Itoa(resp.StatusCode)
		_ = resp.Body.Close()
	}
	if err != nil {
		message = fmt.Sprintf("HTTP check failed: %v", err)
	}
	return PingResult{Timestamp: now, Target: t, Success: success, Latency: latency, Message: message}
}
