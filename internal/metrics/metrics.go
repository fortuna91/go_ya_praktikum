package metrics

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
)

const Gauge = "gauge"
const Counter = "counter"

type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

type Metrics struct {
	values map[string]*Metric
	mtx    sync.RWMutex
}

func (metrics *Metrics) RestoreMetrics(values map[string]*Metric) {
	metrics.values = values
	log.Printf("Metrics were restored: %v\n", values)
}

func (metrics *Metrics) SetGauge(id string, val *float64) {
	metrics.mtx.Lock()
	defer metrics.mtx.Unlock()
	if metrics.values == nil {
		metrics.values = make(map[string]*Metric)
	}
	metrics.values[id] = &Metric{ID: id, MType: Gauge, Value: val}
	log.Printf("Set %v\n", metrics.values[id])
}

func (metrics *Metrics) Get(id string) *Metric {
	metrics.mtx.Lock()
	defer metrics.mtx.Unlock()
	if metrics.values == nil {
		return nil
	}
	log.Printf("Get %v\n", metrics.values[id])
	return metrics.values[id]
}

func (metrics *Metrics) UpdateCounter(id string, val int64) int64 {
	metrics.mtx.Lock()
	defer metrics.mtx.Unlock()
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
	log.Printf("Set %v. Recieved delta %v\n", metrics.values[id], val)
	return *metrics.values[id].Delta
}

func (metrics *Metrics) List() *[]Metric {
	metrics.mtx.Lock()
	defer metrics.mtx.Unlock()
	if metrics.values == nil {
		return &[]Metric{}
	}
	var currMetrics []Metric
	for _, m := range metrics.values {
		currMetrics = append(currMetrics, *m)
	}
	return &currMetrics
}

func CalcHash(metric *Metric, key string) (hash string) {
	hashedString := ""
	if metric.MType == Gauge {
		hashedString = fmt.Sprintf("%s:%s:%f", metric.ID, Gauge, *metric.Value)
	} else if metric.MType == Counter {
		hashedString = fmt.Sprintf("%s:%s:%d", metric.ID, Counter, *metric.Delta)
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(hashedString))
	return hex.EncodeToString(h.Sum(nil))
}

func (metric *Metric) SetHash(key string) {
	metric.Hash = CalcHash(metric, key)
}

func (metric Metric) String() string {
	if metric.MType == Gauge {
		return fmt.Sprintf("id=%s, value=%v", metric.ID, *metric.Value)
	} else if metric.MType == Counter {
		return fmt.Sprintf("id=%s, delta=%v", metric.ID, *metric.Delta)
	}
	return ""
}

// for tests

func (metrics *Metrics) ResetValues() {
	metrics.mtx.Lock()
	defer metrics.mtx.Unlock()
	if metrics.values != nil {
		metrics.values = nil
	}
}
