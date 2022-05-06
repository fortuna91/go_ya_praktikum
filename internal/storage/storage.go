package storage

import (
	"context"
	"github.com/fortuna91/go_ya_praktikum/internal/metrics"
)

type Storage interface {
	Create(ctx context.Context)
	Ping(ctx context.Context) bool
	Restore() map[string]*metrics.Metric
	StoreMetric(context.Context, *metrics.Metric) error
	StoreBatchMetrics(context.Context, []metrics.Metric) error
	Close()
}
