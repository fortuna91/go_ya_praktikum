package storage

import (
	"encoding/json"
	"fmt"
	"github.com/fortuna91/go_ya_praktikum/internal/db"
	"log"
	"os"
	"time"

	"github.com/fortuna91/go_ya_praktikum/internal/configs"
	"github.com/fortuna91/go_ya_praktikum/internal/metrics"
)

type writer struct {
	file    *os.File
	encoder *json.Encoder
}

type reader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewWriter(fileName string) (*writer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return nil, err
	}
	return &writer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *writer) WriteMetrics(metric map[string]*metrics.Metric) error {
	return p.encoder.Encode(&metric)
}

func (p *writer) Close() error {
	return p.file.Close()
}

func NewReader(fileName string) (*reader, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &reader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (p *reader) ReadMetrics() (map[string]*metrics.Metric, error) {
	storedMetric := make(map[string]*metrics.Metric)
	if err := p.decoder.Decode(&storedMetric); err != nil {
		if string(err.Error()) == "EOF" {
			return storedMetric, nil
		}
		return nil, err
	}
	return storedMetric, nil
}

func (p *reader) Close() error {
	return p.file.Close()
}

func StoreMetricsTicker(storeTicker *time.Ticker, handlerMetrics *metrics.Metrics, config configs.ServerConfig) {
	if len(config.StoreFile) > 0 {
		for {
			<-storeTicker.C
			StoreMetrics(handlerMetrics, config.StoreFile)
		}
	} else {
		fmt.Println("Do not store metrics")
	}
}

func StoreMetrics(handlerMetrics *metrics.Metrics, storeFile string) {
	producer, err := NewWriter(storeFile)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	fmt.Println("Store metrics...")
	currMetrics := handlerMetrics.List()
	if err := producer.WriteMetrics(currMetrics); err != nil {
		log.Fatal(err)
	}
}

func Restore(handlerMetrics *metrics.Metrics, config configs.ServerConfig, dbAddress string) {
	producer, err := NewReader(config.StoreFile)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	fmt.Println("Restore metrics...")
	var storedMetrics map[string]*metrics.Metric
	if len(dbAddress) > 0 {
		storedMetrics = db.Restore(dbAddress)
	} else {
		storedMetrics, err = producer.ReadMetrics()
		if err != nil {
			fmt.Println("Error while reading")
			log.Fatal(err)
		}
	}

	handlerMetrics.RestoreMetrics(storedMetrics)
}
