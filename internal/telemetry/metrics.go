package telemetry

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Metrics struct {
	mu       sync.Mutex
	counters map[string]uint64
	gauges   map[string]float64
}

func NewMetrics() *Metrics {
	return &Metrics{
		counters: map[string]uint64{},
		gauges:   map[string]float64{},
	}
}

func (m *Metrics) Inc(name string, labels map[string]string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[metricKey(name, labels)]++
}

func (m *Metrics) Set(name string, labels map[string]string, value float64) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[metricKey(name, labels)] = value
}

func (m *Metrics) PrometheusText() string {
	if m == nil {
		return ""
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	keys := make([]string, 0, len(m.counters)+len(m.gauges))
	for key := range m.counters {
		keys = append(keys, key)
	}
	for key := range m.gauges {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, key := range keys {
		if value, ok := m.counters[key]; ok {
			fmt.Fprintf(&b, "%s %d\n", key, value)
			continue
		}
		fmt.Fprintf(&b, "%s %.0f\n", key, m.gauges[key])
	}
	return b.String()
}

func metricKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, key, escapeLabel(labels[key])))
	}
	return name + "{" + strings.Join(parts, ",") + "}"
}

func escapeLabel(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	return value
}
