package monitor

import (
	"crypto/sha1"
	"encoding/hex"
	"net/url"
	"strings"

	"netwatcher/next/internal/domain"
)

func targetID(value string) string {
	sum := sha1.Sum([]byte(strings.ToLower(strings.TrimSpace(value))))
	return hex.EncodeToString(sum[:6])
}

func DefaultTargets() []domain.Target {
	targets := make([]domain.Target, 0, 3)
	if gateway := detectDefaultGateway(); gateway != "" {
		targets = append(targets, domain.Target{ID: targetID("gateway:" + gateway), Name: "Default Gateway", Host: gateway, Kind: "local", Mode: "ping"})
	}
	targets = append(targets,
		domain.Target{ID: targetID("cloudflare:1.1.1.1"), Name: "Cloudflare", Host: "1.1.1.1", Kind: "internet", Mode: "ping"},
		domain.Target{ID: targetID("google:8.8.8.8"), Name: "Google", Host: "8.8.8.8", Kind: "internet", Mode: "ping"},
	)
	return targets
}

func ParseTarget(raw string) domain.Target {
	raw = strings.TrimSpace(raw)
	t := domain.Target{ID: targetID(raw), Name: "Custom: " + raw, Host: raw, Kind: "internet", Mode: "ping", Custom: true}
	lower := strings.ToLower(raw)
	switch {
	case strings.HasPrefix(lower, "tcp://"):
		host := strings.TrimSpace(raw[len("tcp://"):])
		t.Host = host
		t.Name = "TCP: " + host
		t.Mode = "tcp"
	case strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://"):
		if parsed, err := url.Parse(raw); err == nil && parsed.Host != "" {
			t.Name = strings.ToUpper(parsed.Scheme) + ": " + parsed.Host
			t.Mode = strings.ToLower(parsed.Scheme)
		}
	}
	return t
}

func TargetsFromConfig(custom []string) []domain.Target {
	targets := DefaultTargets()
	seen := map[string]bool{}
	for _, t := range targets {
		seen[t.ID] = true
	}
	for _, raw := range custom {
		t := ParseTarget(raw)
		if strings.TrimSpace(t.Host) == "" || seen[t.ID] {
			continue
		}
		seen[t.ID] = true
		targets = append(targets, t)
	}
	return targets
}
