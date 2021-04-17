package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

var (
	aggregator *metricAggregator
)

func NewAggregator() {
	aggregator = newMetricAggregator()
}

func MeasureElapsed(key string, startTime time.Time) {
	aggregator.setDurationMetric(keyWithGlobalNamespace(key), time.Since(startTime).String())
}

func MeasureProjectElapsed(project *model.Project, key string, startTime time.Time) {
	aggregator.setDurationMetric(keyWithProjectNamespace(project, key), time.Since(startTime).String())
}

func IncrementFrequency(key string) {
	aggregator.incrementFrequencyMetric(keyWithGlobalNamespace(key))
}

func keyWithProjectNamespace(project *model.Project, key string) string {
	// ensure project keys are grouped together
	return fmt.Sprintf("%s/%s", project.ProtoPackage, key)
}

func keyWithGlobalNamespace(key string) string {
	// ensure global keys are grouped together, and listed first in the map
	return fmt.Sprintf("%s/%s", "@code-generator", key)
}

func Flush(writer io.Writer) error {
	byt, err := aggregator.getMetrics()
	if err != nil {
		return err
	}
	_, err = writer.Write(byt)
	return err
}

// This is a primitive implementation for compiling performance measurements of code-gen
// If we need it, we could substitute this with something more heavy handed like:
//	https://github.com/armon/go-metrics
type metricAggregator struct {
	metricsMu sync.Mutex

	DurationMetrics  map[string]string `json:"duration"`
	FrequencyMetrics map[string]int64  `json:"frequency"`
}

func newMetricAggregator() *metricAggregator {
	return &metricAggregator{
		DurationMetrics:  map[string]string{},
		FrequencyMetrics: map[string]int64{},
	}
}

func (m *metricAggregator) setDurationMetric(key, value string) {
	m.metricsMu.Lock()
	defer m.metricsMu.Unlock()
	m.DurationMetrics[key] = value
}

func (m *metricAggregator) incrementFrequencyMetric(key string) {
	m.metricsMu.Lock()
	defer m.metricsMu.Unlock()
	v, ok := m.FrequencyMetrics[key]
	if ok {
		m.FrequencyMetrics[key] = v + 1
	} else {
		m.FrequencyMetrics[key] = 1
	}
}

func (m *metricAggregator) getMetrics() ([]byte, error) {
	m.metricsMu.Lock()
	defer m.metricsMu.Unlock()

	return json.MarshalIndent(m, "", "    ")
}
