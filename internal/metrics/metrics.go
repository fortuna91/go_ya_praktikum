package metrics

import (
	"fmt"
	"strconv"
	"sync"
)

type Metric struct {
	Name  string
	Type  string
	Value string
}

type Metrics struct {
	values map[string]Metric
	sync.RWMutex
}

func (metrics *Metrics) Set(k string, v string) {
	metrics.Lock()
	defer metrics.Unlock()
	if metrics.values == nil {
		metrics.values = make(map[string]Metric)
	}
	metrics.values[k] = Metric{Name: k, Value: v}
}

func (metrics *Metrics) Get(k string) string {
	metrics.Lock()
	defer metrics.Unlock()
	if metrics.values == nil {
		return ""
	}
	return metrics.values[k].Value
}

func (metrics *Metrics) UpdateCounter(k string, v int64) string {
	metrics.Lock()
	defer metrics.Unlock()
	if metrics.values == nil {
		metrics.values = make(map[string]Metric)
		metrics.values[k] = Metric{Name: k, Value: strconv.FormatInt(v, 10)}
	} else if metrics.values[k] == (Metric{}) { // fixme "empty" check
		metrics.values[k] = Metric{Name: k, Value: strconv.FormatInt(v, 10)}
	} else {
		currVal, err := strconv.ParseInt(metrics.values[k].Value, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("Wrong current value of conter metric: %v=%v", k, metrics.values[k]))
		}
		newVal := currVal + v
		metrics.values[k] = Metric{Name: k, Value: strconv.FormatInt(newVal, 10)}
	}
	return metrics.values[k].Value
}

func (metrics *Metrics) List() map[string]Metric {
	metrics.Lock()
	defer metrics.Unlock()
	if metrics.values == nil {
		return map[string]Metric{}
	}
	return metrics.values
}

func (metrics *Metrics) ResetValues() {
	metrics.Lock()
	defer metrics.Unlock()
	if metrics.values != nil {
		metrics.values = nil
	}
}
