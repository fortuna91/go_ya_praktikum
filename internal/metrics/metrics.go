package metrics

import (
	"fmt"
	"sync"
)

const Gauge = "gauge"
const Counter = "counter"

type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type Metrics struct {
	values map[string]*Metric
	sync.RWMutex
}

func (metrics *Metrics) SetGauge(id string, val *float64) {
	metrics.Lock()
	defer metrics.Unlock()
	if metrics.values == nil {
		metrics.values = make(map[string]*Metric)
	}
	metrics.values[id] = &Metric{ID: id, MType: Gauge, Value: val}
	fmt.Printf("SET %v %v\n", id, metrics.values[id])
}

func (metrics *Metrics) Get(id string) *Metric {
	metrics.Lock()
	defer metrics.Unlock()
	if metrics.values == nil {
		return nil
	}
	fmt.Printf("GET %v %v\n", id, metrics.values[id])
	return metrics.values[id]
}

func (metrics *Metrics) UpdateCounter(id string, val int64) int64 {
	metrics.Lock()
	defer metrics.Unlock()
	if metrics.values == nil {
		metrics.values = make(map[string]*Metric)
		metrics.values[id] = &Metric{ID: id, MType: Counter, Delta: &val}
	} else if metrics.values[id] == nil {
		metrics.values[id] = &Metric{ID: id, MType: Counter, Delta: &val}
	} else {
		currVal := *metrics.values[id].Delta
		newVal := currVal + val
		metrics.values[id] = &Metric{ID: id, MType: Counter, Delta: &newVal}
	}
	return *metrics.values[id].Delta
}

func (metrics *Metrics) List() map[string]*Metric {
	metrics.Lock()
	defer metrics.Unlock()
	if metrics.values == nil {
		return map[string]*Metric{}
	}
	fmt.Println(metrics.values)
	return metrics.values
}

// for tests

func (metrics *Metrics) ResetValues() {
	metrics.Lock()
	defer metrics.Unlock()
	if metrics.values != nil {
		metrics.values = nil
	}
}
