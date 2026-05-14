package domain

import "time"

// Event represents a high-frequency action ingested into the system
type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`      // e.g., "order_created", "user_login"
	UserID    string    `json:"user_id"`
	Value     float64   `json:"value"`     // e.g., order amount
	Timestamp time.Time `json:"timestamp"`
}

// Metric represents a computed sliding-window aggregate
type Metric struct {
	WindowStart time.Time `json:"window_start"`
	WindowEnd   time.Time `json:"window_end"`
	EventCount  int       `json:"event_count"`
	TotalValue  float64   `json:"total_value"`
	AverageValue float64   `json:"avg_value"`
}

// StreamService defines the interface for interacting with Redis Streams
type StreamService interface {
	Publish(event Event) error
	Consume(group string, consumer string, handler func(Event) error) error
}

// AnalyticsService defines the interface for metrics computation
type AnalyticsService interface {
	ProcessEvent(event Event)
	GetLiveMetrics() Metric
}
