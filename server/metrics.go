package server

import (
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics tracks application performance and usage metrics
type Metrics struct {
	mu sync.RWMutex

	// Counters
	MessagesReceived int64 `json:"messages_received"`
	MessagesSent     int64 `json:"messages_sent"`
	MessagesDeleted  int64 `json:"messages_deleted"`
	APIRequests      int64 `json:"api_requests"`
	APIErrors        int64 `json:"api_errors"`
	SMTPConnections  int64 `json:"smtp_connections"`
	SMTPErrors       int64 `json:"smtp_errors"`
	AuthAttempts     int64 `json:"auth_attempts"`
	AuthFailures     int64 `json:"auth_failures"`
	RateLimitHits    int64 `json:"rate_limit_hits"`

	// Gauges
	ActiveConnections int64 `json:"active_connections"`
	DatabaseSize      int64 `json:"database_size_bytes"`
	MessageCount      int64 `json:"message_count"`

	// Histograms (simplified - track average and max)
	APILatency      *LatencyTracker `json:"api_latency_ms"`
	SMTPLatency     *LatencyTracker `json:"smtp_latency_ms"`
	DatabaseLatency *LatencyTracker `json:"database_latency_ms"`

	// Metadata
	StartTime     time.Time `json:"start_time"`
	LastResetTime time.Time `json:"last_reset_time"`
}

// LatencyTracker tracks latency statistics
type LatencyTracker struct {
	mu         sync.Mutex
	count      int64
	sum        int64
	max        int64
	min        int64
	samples    []int64 // Keep last N samples for percentiles
	maxSamples int
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	now := time.Now()
	return &Metrics{
		APILatency:      NewLatencyTracker(1000),
		SMTPLatency:     NewLatencyTracker(1000),
		DatabaseLatency: NewLatencyTracker(1000),
		StartTime:       now,
		LastResetTime:   now,
	}
}

// NewLatencyTracker creates a new latency tracker
func NewLatencyTracker(maxSamples int) *LatencyTracker {
	return &LatencyTracker{
		samples:    make([]int64, 0, maxSamples),
		maxSamples: maxSamples,
		min:        int64(^uint64(0) >> 1), // Max int64
	}
}

// Record records a latency measurement
func (lt *LatencyTracker) Record(latencyMs int64) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	lt.count++
	lt.sum += latencyMs

	if latencyMs > lt.max {
		lt.max = latencyMs
	}
	if latencyMs < lt.min {
		lt.min = latencyMs
	}

	// Keep rolling window of samples
	if len(lt.samples) >= lt.maxSamples {
		lt.samples = lt.samples[1:]
	}
	lt.samples = append(lt.samples, latencyMs)
}

// Stats returns latency statistics
func (lt *LatencyTracker) Stats() map[string]interface{} {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	if lt.count == 0 {
		return map[string]interface{}{
			"count": 0,
			"avg":   0,
			"max":   0,
			"min":   0,
			"p50":   0,
			"p95":   0,
			"p99":   0,
		}
	}

	avg := lt.sum / lt.count

	// Calculate percentiles from samples
	p50, p95, p99 := calculatePercentiles(lt.samples)

	return map[string]interface{}{
		"count": lt.count,
		"avg":   avg,
		"max":   lt.max,
		"min":   lt.min,
		"p50":   p50,
		"p95":   p95,
		"p99":   p99,
	}
}

// calculatePercentiles calculates percentiles from samples
func calculatePercentiles(samples []int64) (p50, p95, p99 int64) {
	if len(samples) == 0 {
		return 0, 0, 0
	}

	// Simple percentile calculation (not perfectly accurate but good enough)
	// In production, use a proper percentile algorithm
	sorted := make([]int64, len(samples))
	copy(sorted, samples)

	// Simple bubble sort for small datasets
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	p50Index := len(sorted) * 50 / 100
	p95Index := len(sorted) * 95 / 100
	p99Index := len(sorted) * 99 / 100

	if p50Index < len(sorted) {
		p50 = sorted[p50Index]
	}
	if p95Index < len(sorted) {
		p95 = sorted[p95Index]
	}
	if p99Index < len(sorted) {
		p99 = sorted[p99Index]
	}

	return p50, p95, p99
}

// IncrementCounter atomically increments a counter
func (m *Metrics) IncrementCounter(name string) {
	switch name {
	case "messages_received":
		atomic.AddInt64(&m.MessagesReceived, 1)
	case "messages_sent":
		atomic.AddInt64(&m.MessagesSent, 1)
	case "messages_deleted":
		atomic.AddInt64(&m.MessagesDeleted, 1)
	case "api_requests":
		atomic.AddInt64(&m.APIRequests, 1)
	case "api_errors":
		atomic.AddInt64(&m.APIErrors, 1)
	case "smtp_connections":
		atomic.AddInt64(&m.SMTPConnections, 1)
	case "smtp_errors":
		atomic.AddInt64(&m.SMTPErrors, 1)
	case "auth_attempts":
		atomic.AddInt64(&m.AuthAttempts, 1)
	case "auth_failures":
		atomic.AddInt64(&m.AuthFailures, 1)
	case "rate_limit_hits":
		atomic.AddInt64(&m.RateLimitHits, 1)
	}
}

// UpdateGauge updates a gauge value
func (m *Metrics) UpdateGauge(name string, value int64) {
	switch name {
	case "active_connections":
		atomic.StoreInt64(&m.ActiveConnections, value)
	case "database_size":
		atomic.StoreInt64(&m.DatabaseSize, value)
	case "message_count":
		atomic.StoreInt64(&m.MessageCount, value)
	}
}

// RecordLatency records a latency measurement
func (m *Metrics) RecordLatency(category string, latencyMs int64) {
	switch category {
	case "api":
		m.APILatency.Record(latencyMs)
	case "smtp":
		m.SMTPLatency.Record(latencyMs)
	case "database":
		m.DatabaseLatency.Record(latencyMs)
	}
}

// GetSnapshot returns a snapshot of current metrics
func (m *Metrics) GetSnapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.StartTime)

	return map[string]interface{}{
		"counters": map[string]int64{
			"messages_received": atomic.LoadInt64(&m.MessagesReceived),
			"messages_sent":     atomic.LoadInt64(&m.MessagesSent),
			"messages_deleted":  atomic.LoadInt64(&m.MessagesDeleted),
			"api_requests":      atomic.LoadInt64(&m.APIRequests),
			"api_errors":        atomic.LoadInt64(&m.APIErrors),
			"smtp_connections":  atomic.LoadInt64(&m.SMTPConnections),
			"smtp_errors":       atomic.LoadInt64(&m.SMTPErrors),
			"auth_attempts":     atomic.LoadInt64(&m.AuthAttempts),
			"auth_failures":     atomic.LoadInt64(&m.AuthFailures),
			"rate_limit_hits":   atomic.LoadInt64(&m.RateLimitHits),
		},
		"gauges": map[string]int64{
			"active_connections": atomic.LoadInt64(&m.ActiveConnections),
			"database_size":      atomic.LoadInt64(&m.DatabaseSize),
			"message_count":      atomic.LoadInt64(&m.MessageCount),
		},
		"latency": map[string]interface{}{
			"api":      m.APILatency.Stats(),
			"smtp":     m.SMTPLatency.Stats(),
			"database": m.DatabaseLatency.Stats(),
		},
		"metadata": map[string]interface{}{
			"start_time":      m.StartTime,
			"uptime_seconds":  int64(uptime.Seconds()),
			"uptime_readable": uptime.String(),
			"last_reset":      m.LastResetTime,
		},
	}
}

// Reset resets all metrics (useful for testing)
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Reset counters
	atomic.StoreInt64(&m.MessagesReceived, 0)
	atomic.StoreInt64(&m.MessagesSent, 0)
	atomic.StoreInt64(&m.MessagesDeleted, 0)
	atomic.StoreInt64(&m.APIRequests, 0)
	atomic.StoreInt64(&m.APIErrors, 0)
	atomic.StoreInt64(&m.SMTPConnections, 0)
	atomic.StoreInt64(&m.SMTPErrors, 0)
	atomic.StoreInt64(&m.AuthAttempts, 0)
	atomic.StoreInt64(&m.AuthFailures, 0)
	atomic.StoreInt64(&m.RateLimitHits, 0)

	// Reset latency trackers
	m.APILatency = NewLatencyTracker(1000)
	m.SMTPLatency = NewLatencyTracker(1000)
	m.DatabaseLatency = NewLatencyTracker(1000)

	m.LastResetTime = time.Now()
}

// MetricsHandler returns an HTTP handler for the metrics endpoint
func MetricsHandler(metrics *Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshot := metrics.GetSnapshot()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(snapshot); err != nil {
			http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
		}
	}
}

// MetricsMiddleware records metrics for HTTP requests
func MetricsMiddleware(metrics *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Record API request
			metrics.IncrementCounter("api_requests")

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Record latency
			latency := time.Since(start).Milliseconds()
			metrics.RecordLatency("api", latency)

			// Record errors
			if wrapped.statusCode >= 400 {
				metrics.IncrementCounter("api_errors")
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
