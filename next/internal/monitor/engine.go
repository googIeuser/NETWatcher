package monitor

import (
	"context"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"netwatcher/next/internal/domain"
	"netwatcher/next/internal/storage"
)

const maxHistoryPerTarget = 100000

type Engine struct {
	mu           sync.RWMutex
	config       domain.Config
	targets      []domain.Target
	latest       map[string]domain.Result
	history      map[string][]domain.Sample
	monitoring   bool
	stop         chan struct{}
	samples      int
	outages      []domain.Outage
	activeOutage *domain.Outage
	recentEvents []domain.Event
	currentState string
	currentLabel string
	pendingState string
	pendingLabel string
	pendingCount int
	store        *storage.Store
	version      string
	generation   uint64
	onEvent      func(domain.Event)
}

func NewEngine(cfg domain.Config, store *storage.Store, version string) *Engine {
	e := &Engine{
		config:       cfg,
		targets:      TargetsFromConfig(cfg.CustomTargets),
		latest:       map[string]domain.Result{},
		history:      map[string][]domain.Sample{},
		currentState: "stopped",
		currentLabel: localized(cfg.Language, "Monitoring is stopped", "İzleme durduruldu"),
		store:        store,
		version:      version,
	}
	e.restoreHistory()
	if store != nil {
		_ = store.ApplyRetention(cfg.LogRetentionDays)
	}
	return e
}

func (e *Engine) restoreHistory() {
	if e.store == nil {
		return
	}
	measurements, _ := e.store.ReadMeasurements(time.Now().Add(-24 * time.Hour))
	for _, m := range measurements {
		target := domain.Target{ID: m.TargetID, Name: m.TargetName, Host: m.Host, Kind: m.Kind, Mode: m.Mode}
		for _, current := range e.targets {
			if strings.EqualFold(strings.TrimSpace(current.Host), strings.TrimSpace(m.Host)) || strings.EqualFold(strings.TrimSpace(current.Name), strings.TrimSpace(m.TargetName)) {
				target = current
				break
			}
		}
		e.history[target.ID] = append(e.history[target.ID], domain.Sample{Time: m.Timestamp.Format(time.RFC3339Nano), Latency: m.Latency, Success: m.Success})
		e.latest[target.ID] = domain.Result{Timestamp: m.Timestamp, Target: target, Success: m.Success, Latency: m.Latency, Message: m.Message}
		e.samples++
	}
	e.outages, _ = e.store.ReadOutages(time.Time{})
	e.recentEvents, _ = e.store.ReadEvents(40)
}

func (e *Engine) SetEventHandler(handler func(domain.Event)) {
	e.mu.Lock()
	e.onEvent = handler
	e.mu.Unlock()
}

func (e *Engine) UpdateConfig(cfg domain.Config) {
	e.mu.Lock()
	e.config = cfg
	e.targets = TargetsFromConfig(cfg.CustomTargets)
	e.mu.Unlock()
	if e.store != nil {
		_ = e.store.ApplyRetention(cfg.LogRetentionDays)
	}
}

func (e *Engine) Start() {
	e.mu.Lock()
	if e.monitoring {
		e.mu.Unlock()
		return
	}
	e.monitoring = true
	e.currentState, e.currentLabel = "waiting", e.text("Waiting for the first measurement", "İlk ölçüm bekleniyor")
	e.pendingState, e.pendingLabel, e.pendingCount = "", "", 0
	e.stop = make(chan struct{})
	e.generation++
	generation := e.generation
	stop := e.stop
	event := e.addEventLocked("info", "monitoring", e.text("Monitoring started.", "İzleme başlatıldı."))
	e.mu.Unlock()
	e.dispatchEvent(event)
	go e.loop(stop, generation)
}

func (e *Engine) Stop() {
	e.mu.Lock()
	if !e.monitoring {
		e.mu.Unlock()
		return
	}
	e.monitoring = false
	e.generation++
	stop := e.stop
	e.stop = nil
	ended := e.finishActiveOutageLocked(time.Now(), e.text("Monitoring stopped by the user.", "İzleme kullanıcı tarafından durduruldu."))
	e.currentState, e.currentLabel = "stopped", e.text("Monitoring is stopped", "İzleme durduruldu")
	event := e.addEventLocked("info", "monitoring", e.text("Monitoring stopped.", "İzleme durduruldu."))
	e.mu.Unlock()
	if stop != nil {
		close(stop)
	}
	if ended != nil && e.store != nil {
		_ = e.store.AppendOutage(*ended)
	}
	e.dispatchEvent(event)
}

func (e *Engine) loop(stop <-chan struct{}, generation uint64) {
	e.runCycle(generation)
	for {
		e.mu.RLock()
		interval := e.config.Interval
		e.mu.RUnlock()
		if interval < .5 {
			interval = .5
		}
		timer := time.NewTimer(time.Duration(interval * float64(time.Second)))
		select {
		case <-timer.C:
			e.runCycle(generation)
		case <-stop:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		}
	}
}

func (e *Engine) runCycle(generation uint64) {
	e.mu.RLock()
	targets := append([]domain.Target(nil), e.targets...)
	timeout := time.Duration(e.config.TimeoutMS) * time.Millisecond
	e.mu.RUnlock()
	results := make(chan domain.Result, len(targets))
	var wg sync.WaitGroup
	for _, target := range targets {
		wg.Add(1)
		go func(t domain.Target) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			results <- checkTarget(ctx, t, timeout)
		}(target)
	}
	wg.Wait()
	close(results)
	batch := make([]domain.Result, 0, len(targets))
	for result := range results {
		batch = append(batch, result)
	}
	e.mu.RLock()
	current := e.monitoring && generation == e.generation
	e.mu.RUnlock()
	if !current {
		return
	}
	for _, result := range batch {
		if e.store != nil {
			_ = e.store.AppendMeasurement(domain.Measurement{Timestamp: result.Timestamp, TargetID: result.Target.ID, TargetName: result.Target.Name, Host: result.Target.Host, Kind: result.Target.Kind, Mode: result.Target.Mode, Success: result.Success, Latency: result.Latency, Message: result.Message})
		}
	}

	var emitted []domain.Event
	var completed *domain.Outage
	e.mu.Lock()
	if !e.monitoring || generation != e.generation {
		e.mu.Unlock()
		return
	}
	for _, result := range batch {
		e.latest[result.Target.ID] = result
		h := append(e.history[result.Target.ID], domain.Sample{Time: result.Timestamp.Format(time.RFC3339Nano), Latency: result.Latency, Success: result.Success})
		if len(h) > maxHistoryPerTarget {
			h = append([]domain.Sample(nil), h[len(h)-maxHistoryPerTarget:]...)
		}
		e.history[result.Target.ID] = h
		e.samples++
	}
	rawState, rawLabel := e.rawConnectionStateLocked()
	confirm := e.config.ConfirmCycles
	if confirm < 1 {
		confirm = 1
	}
	if e.currentState == "waiting" || e.currentState == "stopped" {
		e.currentState, e.currentLabel = rawState, rawLabel
		emitted, completed = e.transitionLocked("waiting", rawState, rawLabel)
	} else if rawState == e.currentState {
		e.pendingState, e.pendingLabel, e.pendingCount = "", "", 0
		e.currentLabel = rawLabel
	} else {
		if rawState != e.pendingState {
			e.pendingState, e.pendingLabel, e.pendingCount = rawState, rawLabel, 1
		} else {
			e.pendingCount++
		}
		if e.pendingCount >= confirm {
			old := e.currentState
			e.currentState, e.currentLabel = rawState, rawLabel
			e.pendingState, e.pendingLabel, e.pendingCount = "", "", 0
			emitted, completed = e.transitionLocked(old, rawState, rawLabel)
		}
	}
	e.mu.Unlock()
	if completed != nil && e.store != nil {
		_ = e.store.AppendOutage(*completed)
	}
	for _, event := range emitted {
		e.dispatchEvent(event)
	}
}

func (e *Engine) transitionLocked(oldState, newState, label string) ([]domain.Event, *domain.Outage) {
	now := time.Now()
	problem := isProblemState(newState)
	wasProblem := isProblemState(oldState)
	var events []domain.Event
	var completed *domain.Outage
	if problem && !wasProblem {
		e.activeOutage = &domain.Outage{Start: now.Format(time.RFC3339Nano), Category: newState, Details: label, Active: true}
		event := e.addEventLocked("warning", "outage", e.text("Incident started: ", "Sorun başladı: ")+label)
		events = append(events, event)
	} else if !problem && wasProblem {
		completed = e.finishActiveOutageLocked(now, e.text("Connection returned to normal.", "Bağlantı normale döndü."))
		event := e.addEventLocked("success", "recovery", e.text("Connection returned to normal.", "Bağlantı normale döndü."))
		events = append(events, event)
	} else if problem && wasProblem && oldState != newState {
		completed = e.finishActiveOutageLocked(now, e.text("Incident category changed.", "Sorun kategorisi değişti."))
		e.activeOutage = &domain.Outage{Start: now.Format(time.RFC3339Nano), Category: newState, Details: label, Active: true}
		event := e.addEventLocked("warning", "outage", e.text("Incident changed: ", "Sorun değişti: ")+label)
		events = append(events, event)
	}
	return events, completed
}

func isProblemState(state string) bool {
	return state == "local" || state == "offline" || state == "partial" || state == "degraded"
}

func (e *Engine) finishActiveOutageLocked(now time.Time, details string) *domain.Outage {
	if e.activeOutage == nil {
		return nil
	}
	out := *e.activeOutage
	out.End = now.Format(time.RFC3339Nano)
	out.Active = false
	start, _ := time.Parse(time.RFC3339Nano, out.Start)
	out.DurationSeconds = now.Sub(start).Seconds()
	if details != "" {
		out.Details = out.Details + " " + details
	}
	e.outages = append([]domain.Outage{out}, e.outages...)
	e.activeOutage = nil
	return &out
}

func (e *Engine) addEventLocked(level, category, message string) domain.Event {
	event := domain.Event{Time: time.Now().Format(time.RFC3339Nano), Level: level, Category: category, Message: message}
	e.recentEvents = append([]domain.Event{event}, e.recentEvents...)
	if len(e.recentEvents) > 50 {
		e.recentEvents = e.recentEvents[:50]
	}
	if e.store != nil {
		_ = e.store.AppendEvent(event)
	}
	return event
}
func (e *Engine) dispatchEvent(event domain.Event) {
	e.mu.RLock()
	handler := e.onEvent
	e.mu.RUnlock()
	if handler != nil {
		handler(event)
	}
}

func (e *Engine) Snapshot() domain.Snapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()
	state, label := e.currentState, e.currentLabel
	if !e.monitoring {
		state, label = "stopped", e.text("Monitoring is stopped", "İzleme durduruldu")
	}
	statuses := make([]domain.TargetStatus, 0, len(e.targets))
	allLatencies := []float64{}
	allSuccess, allSamples := 0, 0
	allJitter := []float64{}
	cutoff := time.Now().Add(-time.Duration(e.config.GraphRangeMinutes) * time.Minute)
	for _, target := range e.targets {
		result, ok := e.latest[target.ID]
		rawHistory := e.history[target.ID]
		history := make([]domain.Sample, 0, len(rawHistory))
		for _, sample := range rawHistory {
			t, err := time.Parse(time.RFC3339Nano, sample.Time)
			if err == nil && !t.Before(cutoff) {
				history = append(history, sample)
			}
		}
		status := domain.TargetStatus{Target: target, State: "waiting", History: history}
		if ok {
			status.State = "offline"
			if result.Success {
				status.State = "online"
			}
			status.Latency = result.Latency
			status.LastCheck = result.Timestamp.Format(time.RFC3339Nano)
			status.Message = result.Message
		}
		loss, jit := calculateMetrics(rawHistory)
		status.PacketLoss = loss
		status.Jitter = jit
		statuses = append(statuses, status)
		for _, sample := range rawHistory {
			allSamples++
			if sample.Success {
				allSuccess++
				allLatencies = append(allLatencies, sample.Latency)
			}
		}
		if jit > 0 {
			allJitter = append(allJitter, jit)
		}
	}
	packetLoss := 0.
	if allSamples > 0 {
		packetLoss = 100 * float64(allSamples-allSuccess) / float64(allSamples)
	}
	avg := average(allLatencies)
	jit := average(allJitter)
	quality := qualityScore(state, avg, packetLoss, jit)
	outageCount := len(e.outages)
	if e.activeOutage != nil {
		outageCount++
	}
	events := append([]domain.Event(nil), e.recentEvents...)
	return domain.Snapshot{Monitoring: e.monitoring, ConnectionState: state, ConnectionLabel: label, QualityScore: quality, AverageLatency: avg, PacketLoss: packetLoss, Jitter: jit, Samples: e.samples, Outages: outageCount, Targets: statuses, RecentEvents: events, UpdatedAt: time.Now().Format(time.RFC3339Nano), Version: e.version}
}

func (e *Engine) Outages(days int) []domain.Outage {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if days <= 0 {
		days = 30
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	out := make([]domain.Outage, 0, len(e.outages)+1)
	if e.activeOutage != nil {
		active := *e.activeOutage
		start, _ := time.Parse(time.RFC3339Nano, active.Start)
		active.DurationSeconds = time.Since(start).Seconds()
		out = append(out, active)
	}
	for _, item := range e.outages {
		start, err := time.Parse(time.RFC3339Nano, item.Start)
		if err == nil && !start.Before(cutoff) {
			out = append(out, item)
		}
	}
	return out
}

func (e *Engine) rawConnectionStateLocked() (string, string) {
	localSeen, localOnline := false, false
	internetSeen, internetOnline, internetFailed := false, false, false
	high := false
	for _, target := range e.targets {
		result, ok := e.latest[target.ID]
		if !ok {
			continue
		}
		if target.Kind == "local" {
			localSeen = true
			localOnline = localOnline || result.Success
		} else {
			internetSeen = true
			if result.Success {
				internetOnline = true
			} else {
				internetFailed = true
			}
		}
		if result.Success && result.Latency >= e.config.HighLatencyMS {
			high = true
		}
	}
	if localSeen && !localOnline {
		return "local", e.text("Local network or gateway is unreachable", "Yerel ağ veya modem/ağ geçidine ulaşılamıyor")
	}
	if internetSeen && !internetOnline {
		return "offline", e.text("All internet targets are unreachable", "Tüm internet hedeflerine ulaşılamıyor")
	}
	if internetOnline && internetFailed {
		return "partial", e.text("Some internet targets are unreachable", "Bazı internet hedeflerine ulaşılamıyor")
	}
	if high {
		return "degraded", e.text("Connection is online with high latency", "Bağlantı çevrimiçi ancak gecikme yüksek")
	}
	if len(e.latest) == 0 {
		return "waiting", e.text("Waiting for the first measurement", "İlk ölçüm bekleniyor")
	}
	return "online", e.text("Connection is operating normally", "Bağlantı normal çalışıyor")
}

func calculateMetrics(history []domain.Sample) (float64, float64) {
	if len(history) == 0 {
		return 0, 0
	}
	window := history
	if len(window) > 60 {
		window = window[len(window)-60:]
	}
	success := 0
	lat := []float64{}
	for _, s := range window {
		if s.Success {
			success++
			lat = append(lat, s.Latency)
		}
	}
	loss := 100 * float64(len(window)-success) / float64(len(window))
	if len(lat) < 2 {
		return loss, 0
	}
	diff := []float64{}
	for i := 1; i < len(lat); i++ {
		diff = append(diff, math.Abs(lat[i]-lat[i-1]))
	}
	return loss, average(diff)
}
func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.
	for _, v := range values {
		total += v
	}
	return total / float64(len(values))
}
func qualityScore(state string, latency, loss, jitter float64) int {
	if state == "offline" || state == "local" {
		return 0
	}
	score := 100 - int(math.Min(60, loss*2.5)) - int(math.Min(20, jitter/4))
	if latency > 50 {
		score -= int(math.Min(20, (latency-50)/10))
	}
	if state == "partial" {
		score -= 20
	}
	if state == "degraded" {
		score -= 10
	}
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func SortSamples(samples []domain.Sample) {
	sort.Slice(samples, func(i, j int) bool { return samples[i].Time < samples[j].Time })
}

func localized(language, english, turkish string) string {
	if language == "tr" {
		return turkish
	}
	return english
}
func (e *Engine) text(english, turkish string) string {
	return localized(e.config.Language, english, turkish)
}
