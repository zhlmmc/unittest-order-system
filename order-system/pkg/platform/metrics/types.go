package metrics

import "time"

// MetricType represents the type of metric
type MetricType int

const (
	// Counter is a cumulative metric that only increases
	Counter MetricType = iota
	// Gauge is a metric that can increase and decrease
	Gauge
	// Histogram measures the distribution of values
	Histogram
)

// Labels represents metric labels
type Labels map[string]string

// Metric represents a single metric
type Metric struct {
	Name        string
	Type        MetricType
	Value       float64
	Labels      Labels
	Description string
	Timestamp   time.Time
}

// Collector defines the metrics collection interface
type Collector interface {
	// Counter operations
	IncrementCounter(name string, value float64, labels Labels)
	GetCounter(name string, labels Labels) float64

	// Gauge operations
	SetGauge(name string, value float64, labels Labels)
	GetGauge(name string, labels Labels) float64

	// Histogram operations
	ObserveHistogram(name string, value float64, labels Labels)
	GetHistogram(name string, labels Labels) []float64

	// General operations
	Register(name string, metricType MetricType, description string) error
	Collect() []Metric
}
