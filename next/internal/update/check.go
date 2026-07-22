package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"netwatcher/next/internal/domain"
)

type release struct {
	TagName     string `json:"tag_name"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
}

func Check(ctx context.Context, current string) (domain.UpdateInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/googIeuser/NetWatcher/releases/latest", nil)
	if err != nil {
		return domain.UpdateInfo{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "NetWatcher/"+current)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return domain.UpdateInfo{Checked: true, CurrentVersion: current, Message: err.Error()}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return domain.UpdateInfo{}, fmt.Errorf("GitHub returned HTTP %d", resp.StatusCode)
	}
	var r release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return domain.UpdateInfo{}, err
	}
	latest := strings.TrimPrefix(strings.TrimSpace(r.TagName), "v")
	available := compare(latest, current) > 0
	msg := "NetWatcher is up to date."
	if available {
		msg = "A newer NetWatcher release is available."
	}
	return domain.UpdateInfo{Checked: true, Available: available, CurrentVersion: current, LatestVersion: latest, ReleaseURL: r.HTMLURL, PublishedAt: r.PublishedAt, Message: msg}, nil
}
func compare(a, b string) int {
	pa := parts(a)
	pb := parts(b)
	for i := 0; i < 3; i++ {
		if pa[i] > pb[i] {
			return 1
		}
		if pa[i] < pb[i] {
			return -1
		}
	}
	return 0
}
func parts(v string) [3]int {
	v = strings.SplitN(v, "-", 2)[0]
	p := strings.Split(v, ".")
	var r [3]int
	for i := 0; i < len(p) && i < 3; i++ {
		r[i], _ = strconv.Atoi(p[i])
	}
	return r
}
