//go:build windows

package main

import "testing"

func TestParseTargetSpec(t *testing.T) {
	cases := []struct {
		raw  string
		mode string
	}{
		{"1.1.1.1", checkModePing},
		{"tcp://example.com:443", checkModeTCP},
		{"https://example.com/path", checkModeHTTPS},
		{"http://example.com", checkModeHTTP},
	}
	for _, tc := range cases {
		got := parseTargetSpec(tc.raw)
		if got.Mode != tc.mode {
			t.Fatalf("%q mode=%q want=%q", tc.raw, got.Mode, tc.mode)
		}
	}
}
