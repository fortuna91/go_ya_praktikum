package metrics

import (
	"fmt"
	"github.com/fortuna91/go_ya_praktikum/internal/configs"
	"github.com/fortuna91/go_ya_praktikum/internal/handlers"
	"github.com/fortuna91/go_ya_praktikum/internal/storage"
	"log"
	"sync"
	"time"
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
	mtx    sync.RWMutex
}

func (metrics *Metrics) RestoreMetrics(values *map[string]*Metric) {
	metrics.values = *values
	fmt.Printf("Metrics were restored: %v\n", *values)
}

func (metrics *Metrics) SetGauge(id string, val *float64) {
	metrics.mtx.Lock()
	defer metrics.mtx.Unlock()
	if metrics.values == nil {
		metrics.values = make(map[string]*Metric)
	}
	metrics.values[id] = &Metric{ID: id, MType: Gauge, Value: val}
	fmt.Printf("Set %v = %v\n", id, metrics.values[id])
}

func (metrics *Metrics) Get(id string) *Metric {
	metrics.mtx.Lock()
	defer metrics.mtx.Unlock()
	if metrics.values == nil {
		return nil
	}
	fmt.Printf("Get %v = %v\n", id, metrics.values[id])
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
	return *metrics.values[id].Delta
}

func (metrics *Metrics) List() map[string]*Metric {
	metrics.mtx.Lock()
	defer metrics.mtx.Unlock()
	if metrics.values == nil {
		return map[string]*Metric{}
	}
	fmt.Println(metrics.values)
	return metrics.values
}

// store

func StoreMetricsTicker(storeTicker *time.Ticker, config configs.ServerConfig) {
	if len(config.StoreFile) > 0 {
		for {
			<-storeTicker.C
			StoreMetrics(config.StoreFile)
		}
	} else {
		fmt.Println("Do not store metrics")
	}
}

func StoreMetrics(storeFile string) {
	producer, err := storage.NewWriter(storeFile)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	fmt.Println("Store metrics...")
	currMetrics := handlers.Metrics.List()
	if err := producer.WriteMetrics(&currMetrics); err != nil {
		log.Fatal(err)
	}
}

func Restore(config configs.ServerConfig) {
	producer, err := storage.NewReader(config.StoreFile)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	fmt.Println("Restore metrics...")
	storedMetrics, err := producer.ReadMetrics()
	if err != nil {
		fmt.Println("Error while reading")
		log.Fatal(err)
	}
	handlers.Metrics.RestoreMetrics(storedMetrics)
}

// for tests

func (metrics *Metrics) ResetValues() {
	metrics.mtx.Lock()
	defer metrics.mtx.Unlock()
	if metrics.values != nil {
		metrics.values = nil
	}
}
