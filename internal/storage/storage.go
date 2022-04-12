package storage

import (
	"encoding/json"
	"github.com/fortuna91/go_ya_praktikum/internal/metrics"
	"os"
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

func (p *writer) WriteMetrics(metric *map[string]*metrics.Metric) error {
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

func (p *reader) ReadMetrics() (*map[string]*metrics.Metric, error) {
	storedMetric := make(map[string]*metrics.Metric)
	if err := p.decoder.Decode(&storedMetric); err != nil {
		if string(err.Error()) == "EOF" {
			return &storedMetric, nil
		}
		return nil, err
	}
	return &storedMetric, nil
}

func (p *reader) Close() error {
	return p.file.Close()
}