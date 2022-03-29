package metrics

import (
	"fmt"
	"strconv"
	"sync"
)

type Metrics struct {
	values map[string]string
	lock   sync.RWMutex
}

func (metrics *Metrics) Set(k string, v string) {
	metrics.lock.Lock()
	defer metrics.lock.Unlock()
	if metrics.values == nil {
		metrics.values = make(map[string]string)
	}
	metrics.values[k] = v
}

func (metrics *Metrics) Get(k string) string {
	metrics.lock.Lock()
	defer metrics.lock.Unlock()
	if metrics.values == nil {
		return ""
	}
	return metrics.values[k]
}

func (metrics *Metrics) UpdateCounter(k string, v int64) string {
	metrics.lock.Lock()
	defer metrics.lock.Unlock()
	if metrics.values == nil {
		metrics.values = make(map[string]string)
		metrics.values[k] = strconv.FormatInt(v, 10)
	} else if len(metrics.values[k]) == 0 { // fixme "empty" check
		metrics.values[k] = strconv.FormatInt(v, 10)
	} else {
		currVal, err := strconv.ParseInt(metrics.values[k], 10, 64)
		if err != nil {
			panic(fmt.Sprintf("Wrong current value of conter metric: %v=%v", k, metrics.values[k]))
		}
		newVal := currVal + v
		metrics.values[k] = strconv.FormatInt(newVal, 10)
	}
	return metrics.values[k]
}

func (metrics *Metrics) List() map[string]string {
	metrics.lock.Lock()
	defer metrics.lock.Unlock()
	if metrics.values == nil {
		return map[string]string{}
	}
	return metrics.values
}

func (metrics *Metrics) ResetValues() {
	metrics.lock.Lock()
	defer metrics.lock.Unlock()
	if metrics.values != nil {
		metrics.values = nil
	}
}
