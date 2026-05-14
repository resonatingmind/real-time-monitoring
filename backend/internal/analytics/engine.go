package analytics

import (
	"sync"
	"time"
	"github.com/user/streaming/internal/domain"
)

type Engine struct {
	mu           sync.RWMutex
	events       []domain.Event
	windowSize   time.Duration
	currentMetric domain.Metric
}

func NewEngine(windowSize time.Duration) *Engine {
	e := &Engine{
		windowSize: windowSize,
	}
	// Start a background cleaner to prune old events
	go e.cleaner()
	return e
}

// ProcessEvent adds a new event and updates metrics
func (e *Engine) ProcessEvent(event domain.Event) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.events = append(e.events, event)
	e.calculateMetrics()
}

// cleaner removes events that are outside the sliding window
func (e *Engine) cleaner() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		e.mu.Lock()
		cutoff := time.Now().Add(-e.windowSize)
		
		validIdx := 0
		for i, ev := range e.events {
			if ev.Timestamp.After(cutoff) {
				validIdx = i
				break
			}
			if i == len(e.events)-1 {
				validIdx = len(e.events)
			}
		}
		
		if validIdx > 0 {
			e.events = e.events[validIdx:]
			e.calculateMetrics()
		}
		e.mu.Unlock()
	}
}

func (e *Engine) calculateMetrics() {
	if len(e.events) == 0 {
		e.currentMetric = domain.Metric{}
		return
	}

	var total float64
	for _, ev := range e.events {
		total += ev.Value
	}

	e.currentMetric = domain.Metric{
		WindowStart:  time.Now().Add(-e.windowSize),
		WindowEnd:    time.Now(),
		EventCount:   len(e.events),
		TotalValue:   total,
		AverageValue: total / float64(len(e.events)),
	}
}

func (e *Engine) GetLiveMetrics() domain.Metric {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.currentMetric
}