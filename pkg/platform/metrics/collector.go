package metrics

import (
	"fmt"
	"sync"
	"time"

	"order-system/pkg/infra/config"
)

// defaultCollector implements the Collector interface
type defaultCollector struct {
	mu           sync.RWMutex
	counters     map[string]map[string]float64   // name -> labels -> value
	gauges       map[string]map[string]float64   // name -> labels -> value
	histograms   map[string]map[string][]float64 // name -> labels -> values
	descriptions map[string]string               // name -> description
	types        map[string]MetricType           // name -> type
}

// New creates a new metrics collector
func New(cfg *config.Config) (Collector, error) {
	if !cfg.Metrics.Enabled {
		return nil, fmt.Errorf("metrics are disabled")
	}

	return &defaultCollector{
		counters:     make(map[string]map[string]float64),
		gauges:       make(map[string]map[string]float64),
		histograms:   make(map[string]map[string][]float64),
		descriptions: make(map[string]string),
		types:        make(map[string]MetricType),
	}, nil
}

// Register implements Collector.Register
func (c *defaultCollector) Register(name string, metricType MetricType, description string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.types[name]; exists {
		return fmt.Errorf("metric %s already registered", name)
	}

	c.types[name] = metricType
	c.descriptions[name] = description

	switch metricType {
	case Counter:
		c.counters[name] = make(map[string]float64)
	case Gauge:
		c.gauges[name] = make(map[string]float64)
	case Histogram:
		c.histograms[name] = make(map[string][]float64)
	}

	return nil
}

// IncrementCounter implements Collector.IncrementCounter
func (c *defaultCollector) IncrementCounter(name string, value float64, labels Labels) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.types[name] != Counter {
		return
	}

	key := labelsToString(labels)
	if _, exists := c.counters[name]; !exists {
		c.counters[name] = make(map[string]float64)
	}
	c.counters[name][key] += value
}

// GetCounter implements Collector.GetCounter
func (c *defaultCollector) GetCounter(name string, labels Labels) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.types[name] != Counter {
		return 0
	}

	key := labelsToString(labels)
	return c.counters[name][key]
}

// SetGauge implements Collector.SetGauge
func (c *defaultCollector) SetGauge(name string, value float64, labels Labels) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.types[name] != Gauge {
		return
	}

	key := labelsToString(labels)
	if _, exists := c.gauges[name]; !exists {
		c.gauges[name] = make(map[string]float64)
	}
	c.gauges[name][key] = value
}

// GetGauge implements Collector.GetGauge
func (c *defaultCollector) GetGauge(name string, labels Labels) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.types[name] != Gauge {
		return 0
	}

	key := labelsToString(labels)
	return c.gauges[name][key]
}

// ObserveHistogram implements Collector.ObserveHistogram
func (c *defaultCollector) ObserveHistogram(name string, value float64, labels Labels) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.types[name] != Histogram {
		return
	}

	key := labelsToString(labels)
	if _, exists := c.histograms[name]; !exists {
		c.histograms[name] = make(map[string][]float64)
	}
	c.histograms[name][key] = append(c.histograms[name][key], value)
}

// GetHistogram implements Collector.GetHistogram
func (c *defaultCollector) GetHistogram(name string, labels Labels) []float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.types[name] != Histogram {
		return nil
	}

	key := labelsToString(labels)
	return c.histograms[name][key]
}

// Collect implements Collector.Collect
func (c *defaultCollector) Collect() []Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var metrics []Metric
	now := time.Now()

	// Collect counters
	for name, values := range c.counters {
		for labelKey, value := range values {
			metrics = append(metrics, Metric{
				Name:        name,
				Type:        Counter,
				Value:       value,
				Labels:      stringToLabels(labelKey),
				Description: c.descriptions[name],
				Timestamp:   now,
			})
		}
	}

	// Collect gauges
	for name, values := range c.gauges {
		for labelKey, value := range values {
			metrics = append(metrics, Metric{
				Name:        name,
				Type:        Gauge,
				Value:       value,
				Labels:      stringToLabels(labelKey),
				Description: c.descriptions[name],
				Timestamp:   now,
			})
		}
	}

	// Collect histograms
	for name, values := range c.histograms {
		for labelKey, histogram := range values {
			for _, value := range histogram {
				metrics = append(metrics, Metric{
					Name:        name,
					Type:        Histogram,
					Value:       value,
					Labels:      stringToLabels(labelKey),
					Description: c.descriptions[name],
					Timestamp:   now,
				})
			}
		}
	}

	return metrics
}

// labelsToString converts Labels to a string key
func labelsToString(labels Labels) string {
	if len(labels) == 0 {
		return ""
	}

	var result string
	for k, v := range labels {
		result += fmt.Sprintf("%s=%s;", k, v)
	}
	return result
}

// stringToLabels converts a string key back to Labels
func stringToLabels(s string) Labels {
	if s == "" {
		return Labels{}
	}

	labels := make(Labels)
	// Simple implementation - in real code, you'd want to properly parse the string
	return labels
}
