package fsStorage

import (
	"context"
	"encoding/json"
	"github.com/fortuna91/go_ya_praktikum/internal/storage"
	"log"
	"os"
	"time"

	"github.com/fortuna91/go_ya_praktikum/internal/metrics"
)

type fileStorage struct {
	fileName string
}

type writer struct {
	file    *os.File
	encoder *json.Encoder
}

type reader struct {
	file    *os.File
	decoder *json.Decoder
}

func (p *writer) WriteMetrics(metric []metrics.Metric) error {
	return p.encoder.Encode(&metric)
}

func (p *reader) ReadMetrics() ([]metrics.Metric, error) {
	var storedMetric []metrics.Metric
	if err := p.decoder.Decode(&storedMetric); err != nil {
		if string(err.Error()) == "EOF" {
			return storedMetric, nil
		}
		return nil, err
	}
	return storedMetric, nil
}

func (p *writer) Close() error {
	return p.file.Close()
}

func (p *reader) Close() error {
	return p.file.Close()
}

func New(fileName string) *fileStorage {
	return &fileStorage{
		fileName: fileName,
	}
}

func (fs *fileStorage) StoreBatchMetrics(_ context.Context, currMetrics []metrics.Metric) error {
	producer, err := newWriter(fs.fileName)
	if err != nil {
		return err
	}
	defer producer.Close()

	log.Println("Store metrics into a file...")
	if err := producer.WriteMetrics(currMetrics); err != nil {
		return err
	}
	return nil
}

func (fs *fileStorage) Restore() map[string]*metrics.Metric {
	producer, err := newReader(fs.fileName)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer producer.Close()

	log.Println("Restore metrics from a file...")
	metricsList, err := producer.ReadMetrics()
	if err != nil {
		log.Fatalf("Error while reading: %v", err)
	}
	restoredMetrics := make(map[string]*metrics.Metric)
	for i := range metricsList {
		restoredMetrics[metricsList[i].ID] = &metricsList[i]
	}

	return restoredMetrics
}

// mocks

func (fs *fileStorage) Close()                                                 {}
func (fs *fileStorage) Create(_ context.Context)                               {}
func (fs *fileStorage) StoreMetric(_ context.Context, _ *metrics.Metric) error { return nil }
func (fs *fileStorage) Ping(_ context.Context) bool                            { return true }

//StoreBatchMetrics

func newWriter(fileName string) (*writer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return nil, err
	}
	return &writer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func newReader(fileName string) (*reader, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &reader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func StoreMetricsTicker(fs *storage.Storage, storeTicker *time.Ticker, handlerMetrics *metrics.Metrics) {
	for {
		<-storeTicker.C
		(*fs).StoreBatchMetrics(context.Background(), *handlerMetrics.List())
	}

}
