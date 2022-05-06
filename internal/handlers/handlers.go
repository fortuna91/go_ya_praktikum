package handlers

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/fortuna91/go_ya_praktikum/internal/storage"
	"github.com/go-chi/chi/v5"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/fortuna91/go_ya_praktikum/internal/metrics"
)

var Metrics = metrics.Metrics{}

// fixme: Do better
var HashKey string
var StoreMetrics = false
var StoreMetricImmediately = false
var Storage storage.Storage

// fixme maybe for feature it has to be channel with mutex
// var CountChannel = make(chan int64)

func SetGaugeMetric(w http.ResponseWriter, r *http.Request) {
	log.Println("Got it...")
	val := chi.URLParam(r, "value")
	if val == "" {
		http.Error(w, "Empty value", http.StatusNotFound)
		return
	}
	floatVal, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Printf("Parse error: %v\n", err)
		http.Error(w, "Wrong metric value", http.StatusBadRequest)
		return
	}
	metricName := chi.URLParam(r, "metricName")
	Metrics.SetGauge(metricName, &floatVal)
	w.WriteHeader(http.StatusOK)
}

func SetCounterMetric(w http.ResponseWriter, r *http.Request) {
	// go Counter(CountChannel)
	val := chi.URLParam(r, "value")
	if val == "" {
		http.Error(w, "Empty value", http.StatusNotFound)
		return
	}
	countVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		log.Printf("Parse error: %v\n", err)
		http.Error(w, "Wrong metric value", http.StatusBadRequest)
		return
	}
	metricName := chi.URLParam(r, "metricName")
	Metrics.UpdateCounter(metricName, countVal)
	w.WriteHeader(http.StatusOK)
}

func GetMetric(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "metricName")
	metric := Metrics.Get(metricName)
	if metric != nil {
		var val = ""
		if metric.MType == metrics.Gauge {
			val = strconv.FormatFloat(*metric.Value, 'f', 3, 64)
		} else if metric.MType == metrics.Counter {
			val = strconv.FormatInt(*metric.Delta, 10)
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(val))
		if err != nil {
			log.Printf("Error sending the response: %v\n", err)
			http.Error(w, "Error sending the response", http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func ListMetrics(w http.ResponseWriter, _ *http.Request) {
	const tmplHTML = `<!DOCTYPE html>
	<html>
	<table>
	<tr>
	<td>Name</td>
	<td>Value</td>
	</tr>{{range . }}
	<tr>
	<td>{{ .ID }}</td>
	<td>{{ .MType }}</td>
	<td>{{ .Delta }}</td>
	<td>{{ .Value }}</td>
	</tr>{{end}}
	</table>
	</html>`

	listMetrics := Metrics.List()

	w.Header().Set("Content-Type", "text/html")
	tmpl, err := template.New("index").Parse(tmplHTML)
	if err != nil {
		http.Error(w, "Error getting the template", http.StatusInternalServerError)
	}
	errEx := tmpl.Execute(w, listMetrics)
	if errEx != nil {
		log.Printf("Error sending the response: %v\n", errEx)
	}
}

func NotImplemented(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// JSON

func getReader(w http.ResponseWriter, r *http.Request) *io.ReadCloser {
	var reader io.ReadCloser
	var err error

	if r.Header.Get(`Content-Encoding`) == `gzip` {
		reader, err = gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
	} else {
		reader = r.Body
	}
	return &reader
}

func SetMetricJSON(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	reader := *getReader(w, r)
	if reader == nil {
		http.Error(w, "Couldn't get reader", http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	respBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Couldn't read body %v\n", err)
		http.Error(w, "Couldn't read body", http.StatusInternalServerError)
		return
	}
	metricRequest := metrics.Metric{}
	json.Unmarshal(respBody, &metricRequest)
	if len(metricRequest.ID) == 0 {
		http.Error(w, "Empty metric id", http.StatusBadRequest)
		return
	}

	if len(HashKey) > 0 {
		metricHash := metrics.CalcHash(&metricRequest, HashKey)
		if metricHash != metricRequest.Hash {
			log.Printf("Incorrect data hash: %s != %s\n", metricRequest.Hash, metricHash)
			http.Error(w, "Incorrect data hash", http.StatusBadRequest)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	var newMetric metrics.Metric
	if metricRequest.MType == metrics.Gauge {
		if metricRequest.Value == nil {
			http.Error(w, "Empty metric value", http.StatusBadRequest)
			return
		}
		Metrics.SetGauge(metricRequest.ID, metricRequest.Value)

		if StoreMetrics {
			if StoreMetricImmediately {
				Storage.StoreBatchMetrics(ctx, *Metrics.List())
			} else if err := Storage.StoreMetric(ctx, &metricRequest); err != nil {
				http.Error(w, "Couldn't store metric", http.StatusInternalServerError)
			}
		}
		newMetric = metricRequest
		w.WriteHeader(http.StatusOK)
	} else if metricRequest.MType == metrics.Counter {
		if metricRequest.Delta == nil {
			http.Error(w, "Empty metric delta", http.StatusBadRequest)
			return
		}

		newDelta := Metrics.UpdateCounter(metricRequest.ID, *metricRequest.Delta)
		newMetric = metrics.Metric{ID: metricRequest.ID, MType: metricRequest.MType, Delta: &newDelta}
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
		return
	}

	if StoreMetrics {
		if StoreMetricImmediately {
			Storage.StoreBatchMetrics(ctx, *Metrics.List())
		} else if err := Storage.StoreMetric(ctx, &newMetric); err != nil {
			http.Error(w, "Couldn't store metric", http.StatusInternalServerError)
		}
	}

	// ??
	metric := metrics.Metric{}
	bodyResp, _ := json.Marshal(metric)
	_, errBody := w.Write(bodyResp)
	if errBody != nil {
		log.Printf("Error sending the response: %v\n", errBody)
		http.Error(w, "Error sending the response", http.StatusInternalServerError)
	}
}

func GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	reader := *getReader(w, r)
	if reader == nil {
		http.Error(w, "Couldn't get reader", http.StatusInternalServerError)
		return
	}
	defer reader.Close()
	respBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Couldn't read body %v\n", err)
		http.Error(w, "Couldn't read body", http.StatusInternalServerError)
		return
	}
	metricRequest := metrics.Metric{}
	json.Unmarshal(respBody, &metricRequest)
	if len(metricRequest.ID) == 0 {
		http.Error(w, "Empty metric ID", http.StatusBadRequest)
		return
	}
	metric := Metrics.Get(metricRequest.ID)
	if metric != nil {
		if len(HashKey) > 0 {
			metric.SetHash(HashKey)
		}
		bodyResp, err := json.Marshal(metric)
		if err != nil {
			log.Printf("Cannot convert Metric to JSON: %v", err)
			http.Error(w, "Error sending the response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, errBody := w.Write(bodyResp)
		if errBody != nil {
			log.Printf("Error sending the response: %v\n", errBody)
			http.Error(w, "Error sending the response", http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func SetBatchMetrics(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	reader := *getReader(w, r)
	if reader == nil {
		http.Error(w, "Couldn't get reader", http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	respBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Couldn't read body %v\n", err)
		http.Error(w, "Couldn't read body", http.StatusInternalServerError)
		return
	}
	var metricsRequest []metrics.Metric
	json.Unmarshal(respBody, &metricsRequest)

	if len(HashKey) > 0 {
		for _, metricRequest := range metricsRequest {
			metricHash := metrics.CalcHash(&metricRequest, HashKey)
			if metricHash != metricRequest.Hash {
				log.Printf("Incorrect data hash: %s != %s\n", metricRequest.Hash, metricHash)
				http.Error(w, "Incorrect data hash", http.StatusBadRequest)
				return
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")

	var updatedMetricsRequest []metrics.Metric
	for _, metricRequest := range metricsRequest {
		if metricRequest.MType == metrics.Gauge {
			if metricRequest.Value == nil {
				http.Error(w, "Empty metric value", http.StatusBadRequest)
				return
			}
			Metrics.SetGauge(metricRequest.ID, metricRequest.Value)

			updatedMetricsRequest = append(updatedMetricsRequest, metricRequest)
		} else if metricRequest.MType == metrics.Counter {
			if metricRequest.Delta == nil {
				http.Error(w, "Empty metric delta", http.StatusBadRequest)
				return
			}

			newDelta := Metrics.UpdateCounter(metricRequest.ID, *metricRequest.Delta)
			metricRequest.Delta = &newDelta

			updatedMetricsRequest = append(updatedMetricsRequest, metricRequest)
		} else {
			http.Error(w, "Unknown metric type", http.StatusBadRequest)
			return
		}
	}

	if StoreMetrics {
		if err = Storage.StoreBatchMetrics(ctx, updatedMetricsRequest); err != nil {
			log.Println(err)
			http.Error(w, "Couldn't set metric into DB", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)

	// ??
	metric := metrics.Metric{}
	bodyResp, _ := json.Marshal(metric)
	if _, errBody := w.Write(bodyResp); errBody != nil {
		log.Printf("Error sending the response: %v\n", errBody)
		http.Error(w, "Error sending the response", http.StatusInternalServerError)
	}
}

func PingDB(w http.ResponseWriter, r *http.Request) {
	res := Storage.Ping(r.Context())
	if res {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "DB is unavailable", http.StatusInternalServerError)
	}
}
