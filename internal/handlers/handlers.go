package handlers

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/fortuna91/go_ya_praktikum/internal/db"
	"github.com/go-chi/chi/v5"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"

	"github.com/fortuna91/go_ya_praktikum/internal/metrics"
	"github.com/fortuna91/go_ya_praktikum/internal/storage"
)

var Metrics = metrics.Metrics{}

// fixme: Do better
var StoreMetricImmediately = true
var StoreFile string
var HashKey string
var DBAddress string

// fixme maybe for feature it has to be channel with mutex
// var CountChannel = make(chan int64)

func SetGaugeMetric(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got it...")
	val := chi.URLParam(r, "value")
	if val == "" {
		http.Error(w, "Empty value", http.StatusNotFound)
		return
	}
	floatVal, err := strconv.ParseFloat(val, 64)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
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
		fmt.Printf("Parse error: %v\n", err)
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
			fmt.Printf("Error sending the response: %v\n", err)
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

	dictMetrics := Metrics.List()
	metricKeys := make([]string, 0, len(dictMetrics))
	for k := range dictMetrics {
		metricKeys = append(metricKeys, k)
	}
	sort.Strings(metricKeys)
	listMetrics := make([]metrics.Metric, 0, len(dictMetrics))
	for _, k := range metricKeys {
		listMetrics = append(listMetrics, *dictMetrics[k])
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl, err := template.New("index").Parse(tmplHTML)
	if err != nil {
		http.Error(w, "Error getting the template", http.StatusInternalServerError)
	}
	errEx := tmpl.Execute(w, listMetrics)
	if errEx != nil {
		fmt.Printf("Error sending the response: %v\n", errEx)
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
	reader := *getReader(w, r)
	if reader == nil {
		http.Error(w, "Couldn't get reader", http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	respBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Couldn't read body %v\n", err)
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
			fmt.Printf("Incorrect data hash: %s != %s\n", metricRequest.Hash, metricHash)
			http.Error(w, "Incorrect data hash", http.StatusBadRequest)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if metricRequest.MType == metrics.Gauge {
		if metricRequest.Value == nil {
			http.Error(w, "Empty metric value", http.StatusBadRequest)
			return
		}
		Metrics.SetGauge(metricRequest.ID, metricRequest.Value)
		w.WriteHeader(http.StatusOK)
	} else if metricRequest.MType == metrics.Counter {
		if metricRequest.Delta == nil {
			http.Error(w, "Empty metric delta", http.StatusBadRequest)
			return
		}
		Metrics.UpdateCounter(metricRequest.ID, *metricRequest.Delta)
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
		return
	}

	if StoreMetricImmediately && len(StoreFile) > 0 {
		storage.StoreMetrics(&Metrics, StoreFile)
	}
	// ??
	metric := metrics.Metric{}
	bodyResp, _ := json.Marshal(metric)
	_, errBody := w.Write(bodyResp)
	if errBody != nil {
		fmt.Printf("Error sending the response: %v\n", errBody)
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
		fmt.Printf("Couldn't read body %v\n", err)
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
			fmt.Printf("Cannot convert Metric to JSON: %v", err)
			http.Error(w, "Error sending the response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, errBody := w.Write(bodyResp)
		if errBody != nil {
			fmt.Printf("Error sending the response: %v\n", errBody)
			http.Error(w, "Error sending the response", http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func PingDB(w http.ResponseWriter, r *http.Request) {
	dbConn := db.Connect(DBAddress)
	res := db.Ping(dbConn)
	if res {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "DB is unavailable", http.StatusInternalServerError)
	}
}
