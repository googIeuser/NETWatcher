package storage

import (
	"archive/zip"
	"crypto/sha1"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"netwatcher/next/internal/domain"
)

type Store struct {
	mu  sync.Mutex
	dir string
}

func New() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return NewAt(filepath.Join(home, "Documents", "NetWatcherLogs"))
}

func NewAt(dir string) (*Store, error) {
	if err := os.MkdirAll(filepath.Join(dir, "Reports"), 0o755); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

func (s *Store) Dir() string        { return s.dir }
func (s *Store) ReportsDir() string { return filepath.Join(s.dir, "Reports") }

func (s *Store) measurementPath(t time.Time) string {
	return filepath.Join(s.dir, "measurements_"+t.Format("2006-01-02")+".csv")
}

// NetWatcher 2.x already owns outages.csv and uses a semicolon-delimited schema.
// Version 3 writes to a separate file so an older installation is never corrupted.
func (s *Store) outagesPath() string       { return filepath.Join(s.dir, "outages_v3.csv") }
func (s *Store) legacyOutagesPath() string { return filepath.Join(s.dir, "outages.csv") }
func (s *Store) eventsPath() string        { return filepath.Join(s.dir, "events_v3.csv") }

func appendCSV(path string, delimiter rune, header, row []string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	w := csv.NewWriter(file)
	w.Comma = delimiter
	if info.Size() == 0 {
		if err := w.Write(header); err != nil {
			return err
		}
	}
	if err := w.Write(row); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

func (s *Store) AppendMeasurement(m domain.Measurement) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return appendCSV(s.measurementPath(m.Timestamp), ',', []string{"timestamp", "target_id", "target_name", "host", "kind", "mode", "success", "latency_ms", "message"}, []string{m.Timestamp.Format(time.RFC3339Nano), m.TargetID, m.TargetName, m.Host, m.Kind, m.Mode, strconv.FormatBool(m.Success), strconv.FormatFloat(m.Latency, 'f', 3, 64), m.Message})
}

func (s *Store) AppendOutage(o domain.Outage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return appendCSV(s.outagesPath(), ',', []string{"start", "end", "category", "details", "duration_seconds"}, []string{o.Start, o.End, o.Category, o.Details, strconv.FormatFloat(o.DurationSeconds, 'f', 3, 64)})
}

func (s *Store) AppendEvent(e domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return appendCSV(s.eventsPath(), ',', []string{"time", "level", "category", "message"}, []string{e.Time, e.Level, e.Category, e.Message})
}

func readCSVFile(path string) (header []string, rows [][]string, err error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	if len(data) == 0 {
		return nil, nil, nil
	}
	firstLine := string(data)
	if i := strings.IndexByte(firstLine, '\n'); i >= 0 {
		firstLine = firstLine[:i]
	}
	delimiter := ','
	if strings.Count(firstLine, ";") > strings.Count(firstLine, ",") {
		delimiter = ';'
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	r.Comma = delimiter
	r.FieldsPerRecord = -1
	all, err := r.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(all) == 0 {
		return nil, nil, nil
	}
	all[0][0] = strings.TrimPrefix(all[0][0], "\ufeff")
	for i := range all[0] {
		all[0][i] = strings.TrimSpace(strings.ToLower(all[0][i]))
	}
	return all[0], all[1:], nil
}

func headerIndex(header []string) map[string]int {
	result := make(map[string]int, len(header))
	for i, value := range header {
		result[strings.TrimSpace(strings.ToLower(value))] = i
	}
	return result
}

func cell(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func parseTime(value string) (time.Time, error) {
	value = strings.TrimSpace(strings.TrimPrefix(value, "\ufeff"))
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported timestamp %q", value)
}

func legacyID(name, host string) string {
	sum := sha1.Sum([]byte(strings.ToLower(strings.TrimSpace(name + "|" + host))))
	return "legacy-" + hex.EncodeToString(sum[:6])
}

func inferLegacyMode(name, host string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	host = strings.ToLower(strings.TrimSpace(host))
	switch {
	case strings.HasPrefix(name, "tcp:") || strings.HasPrefix(host, "tcp://"):
		return "tcp"
	case strings.HasPrefix(name, "https:") || strings.HasPrefix(host, "https://"):
		return "https"
	case strings.HasPrefix(name, "http:") || strings.HasPrefix(host, "http://"):
		return "http"
	default:
		return "ping"
	}
}

func readCurrentMeasurements(path string, since time.Time) ([]domain.Measurement, error) {
	header, rows, err := readCSVFile(path)
	if err != nil || len(header) == 0 {
		return nil, err
	}
	idx := headerIndex(header)
	required := []string{"timestamp", "target_id", "target_name", "host", "kind", "mode", "success", "latency_ms", "message"}
	for _, key := range required {
		if _, ok := idx[key]; !ok {
			return nil, nil
		}
	}
	result := make([]domain.Measurement, 0, len(rows))
	for _, row := range rows {
		ts, err := parseTime(cell(row, idx["timestamp"]))
		if err != nil || ts.Before(since) {
			continue
		}
		success, _ := strconv.ParseBool(cell(row, idx["success"]))
		latency, _ := strconv.ParseFloat(strings.ReplaceAll(cell(row, idx["latency_ms"]), ",", "."), 64)
		result = append(result, domain.Measurement{
			Timestamp: ts, TargetID: cell(row, idx["target_id"]), TargetName: cell(row, idx["target_name"]),
			Host: cell(row, idx["host"]), Kind: cell(row, idx["kind"]), Mode: cell(row, idx["mode"]),
			Success: success, Latency: latency, Message: cell(row, idx["message"]),
		})
	}
	return result, nil
}

func readLegacyMeasurements(path string, since time.Time) ([]domain.Measurement, error) {
	header, rows, err := readCSVFile(path)
	if err != nil || len(header) == 0 {
		return nil, err
	}
	idx := headerIndex(header)
	for _, key := range []string{"timestamp", "name", "host", "target_type", "success", "latency_ms", "message"} {
		if _, ok := idx[key]; !ok {
			return nil, nil
		}
	}
	result := make([]domain.Measurement, 0, len(rows))
	for _, row := range rows {
		ts, err := parseTime(cell(row, idx["timestamp"]))
		if err != nil || ts.Before(since) {
			continue
		}
		name := cell(row, idx["name"])
		host := cell(row, idx["host"])
		success, _ := strconv.ParseBool(cell(row, idx["success"]))
		latency, _ := strconv.ParseFloat(strings.ReplaceAll(cell(row, idx["latency_ms"]), ",", "."), 64)
		result = append(result, domain.Measurement{
			Timestamp: ts, TargetID: legacyID(name, host), TargetName: name, Host: host,
			Kind: cell(row, idx["target_type"]), Mode: inferLegacyMode(name, host),
			Success: success, Latency: latency, Message: cell(row, idx["message"]),
		})
	}
	return result, nil
}

func (s *Store) ReadMeasurements(since time.Time) ([]domain.Measurement, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var result []domain.Measurement
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".csv") {
			continue
		}
		path := filepath.Join(s.dir, entry.Name())
		name := strings.ToLower(entry.Name())
		var items []domain.Measurement
		switch {
		case strings.HasPrefix(name, "measurements_"):
			items, err = readCurrentMeasurements(path, since)
		case strings.HasPrefix(name, "samples_"):
			items, err = readLegacyMeasurements(path, since)
		default:
			continue
		}
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Timestamp.Before(result[j].Timestamp) })
	return result, nil
}

func normalizeOutageCategory(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "LOCAL_NETWORK", "LOCAL":
		return "local"
	case "ISP_OUTAGE", "OFFLINE":
		return "offline"
	case "DEGRADED", "PARTIAL":
		return "partial"
	case "HIGH_LATENCY":
		return "degraded"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func readOutageFile(path string, since time.Time) ([]domain.Outage, error) {
	header, rows, err := readCSVFile(path)
	if err != nil || len(header) == 0 {
		return nil, err
	}
	idx := headerIndex(header)
	startIndex, startOK := idx["start"]
	endIndex, endOK := idx["end"]
	categoryIndex, categoryOK := idx["category"]
	detailsIndex, detailsOK := idx["details"]
	durationIndex, durationOK := idx["duration_seconds"]
	if !startOK || !endOK || !categoryOK || !detailsOK || !durationOK {
		return nil, nil
	}
	out := make([]domain.Outage, 0, len(rows))
	for _, row := range rows {
		start, err := parseTime(cell(row, startIndex))
		if err != nil {
			continue
		}
		endText := cell(row, endIndex)
		if !since.IsZero() {
			end, endErr := parseTime(endText)
			if endErr == nil && end.Before(since) {
				continue
			}
			if endErr != nil && start.Before(since) {
				continue
			}
		}
		dur, _ := strconv.ParseFloat(strings.ReplaceAll(cell(row, durationIndex), ",", "."), 64)
		if dur <= 0 && endText != "" {
			if end, err := parseTime(endText); err == nil {
				dur = end.Sub(start).Seconds()
			}
		}
		out = append(out, domain.Outage{
			Start: start.Format(time.RFC3339Nano), End: endText,
			Category: normalizeOutageCategory(cell(row, categoryIndex)), Details: cell(row, detailsIndex),
			DurationSeconds: dur,
		})
	}
	return out, nil
}

func (s *Store) ReadOutages(since time.Time) ([]domain.Outage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result []domain.Outage
	seen := map[string]bool{}
	for _, path := range []string{s.legacyOutagesPath(), s.outagesPath()} {
		items, err := readOutageFile(path, since)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			key := item.Start + "|" + item.End + "|" + item.Category
			if seen[key] {
				continue
			}
			seen[key] = true
			result = append(result, item)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Start > result[j].Start })
	return result, nil
}

func (s *Store) ReadEvents(limit int) ([]domain.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	header, rows, err := readCSVFile(s.eventsPath())
	if err != nil || len(header) == 0 {
		return nil, err
	}
	idx := headerIndex(header)
	timeIndex, timeOK := idx["time"]
	levelIndex, levelOK := idx["level"]
	categoryIndex, categoryOK := idx["category"]
	messageIndex, messageOK := idx["message"]
	if !timeOK || !levelOK || !categoryOK || !messageOK {
		return nil, nil
	}
	if limit > 0 && len(rows) > limit {
		rows = rows[len(rows)-limit:]
	}
	out := make([]domain.Event, 0, len(rows))
	for i := len(rows) - 1; i >= 0; i-- {
		row := rows[i]
		out = append(out, domain.Event{Time: cell(row, timeIndex), Level: cell(row, levelIndex), Category: cell(row, categoryIndex), Message: cell(row, messageIndex)})
	}
	return out, nil
}

func (s *Store) ApplyRetention(days int) error {
	if days <= 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cutoff := time.Now().AddDate(0, 0, -days)
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		prefix := ""
		switch {
		case strings.HasPrefix(name, "measurements_"):
			prefix = "measurements_"
		case strings.HasPrefix(name, "samples_"):
			prefix = "samples_"
		default:
			continue
		}
		dateText := strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".csv")
		d, err := time.Parse("2006-01-02", dateText)
		if err == nil && d.Before(cutoff) {
			_ = os.Remove(filepath.Join(s.dir, entry.Name()))
		}
	}
	return nil
}

func (s *Store) WriteJSONReport(name string, value any) (string, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	path := filepath.Join(s.ReportsDir(), name)
	return path, os.WriteFile(path, data, 0o644)
}

func (s *Store) ExportZIP(name string, extras map[string][]byte) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := filepath.Join(s.ReportsDir(), name)
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	zw := zip.NewWriter(file)
	closeAll := func(e error) (string, error) {
		_ = zw.Close()
		_ = file.Close()
		if e != nil {
			_ = os.Remove(path)
		}
		return path, e
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return closeAll(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		src := filepath.Join(s.dir, entry.Name())
		in, err := os.Open(src)
		if err != nil {
			continue
		}
		w, err := zw.Create(filepath.ToSlash(filepath.Join("logs", entry.Name())))
		if err == nil {
			_, err = io.Copy(w, in)
		}
		_ = in.Close()
		if err != nil {
			return closeAll(err)
		}
	}
	for name, data := range extras {
		w, err := zw.Create(filepath.ToSlash(name))
		if err != nil {
			return closeAll(err)
		}
		if _, err := w.Write(data); err != nil {
			return closeAll(err)
		}
	}
	if err := zw.Close(); err != nil {
		_ = file.Close()
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	return path, nil
}

func SafeFileName(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, time.Now().Format("20060102_150405"))
}
