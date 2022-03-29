package handlers

import (
	"fmt"
	"github.com/fortuna91/go_ya_praktikum/internal/metrics"
	"github.com/go-chi/chi/v5"
	"net/http"
	"sort"
	"strconv"
)

var Metrics = metrics.Metrics{}

var CurrCount int64 = 0

// fixme maybe for feature it has to be channel with mutex
// var CountChannel = make(chan int64)

/*func Counter(c <-chan int64) {
	for v := range c {
		fmt.Println("here")
		currCount += v
	}
}*/

func Counter(c int64) {
	CurrCount += c
	fmt.Printf("Count is %d\n", CurrCount)
}

func SetGaugeMetric(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got it...")
	val := chi.URLParam(r, "value")
	/*if !contains(gaugeMetrics, items[3]) {
		http.Error(w, "Unknown metric", http.StatusBadRequest)
		return
	}*/
	/*if len(items) < 5 {
		fmt.Printf("Not enough value. Path: %v\n", path)
		http.Error(w, "Not enough value", http.StatusNotFound)
		return
	}*/
	if val == "" {
		http.Error(w, "Empty value", http.StatusNotFound)
		return
	}
	_, err := strconv.ParseFloat(val, 64)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		http.Error(w, "Wrong metric value", http.StatusBadRequest)
		return
	}
	metricType := chi.URLParam(r, "metricName")
	Metrics.Set(metricType, val)
	w.WriteHeader(http.StatusOK)
}

func SetCounterMetric(w http.ResponseWriter, r *http.Request) {
	// go Counter(CountChannel)
	val := chi.URLParam(r, "value")
	/*if items[3] != counterMetric {
		http.Error(w, "Unknown metric", http.StatusBadRequest)
		return
	}*/
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
	metricType := chi.URLParam(r, "metricName")
	Metrics.UpdateCounter(metricType, countVal)
	w.WriteHeader(http.StatusOK)
}

func GetMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricName")
	val := Metrics.Get(metricType)
	if len(val) > 0 { // fixme "empty" check
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(val))
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func ListMetrics(w http.ResponseWriter, _ *http.Request) {
	var form = `<html>
			<table>
  				<tr>
    				<td>Name</td>
    				<td>Value</td>
  				</tr>
				%v
			</table>
		</html>`
	var item = `
		<tr>
			<td>%v</td>
			<td>%v</td>
		</tr>`
	listMetrics := Metrics.List()
	metricKeys := make([]string, 0, len(listMetrics))
	for k := range listMetrics {
		metricKeys = append(metricKeys, k)
	}
	sort.Strings(metricKeys)
	var s = ""
	for _, key := range metricKeys {
		s = s + fmt.Sprintf(item, key, listMetrics[key])
	}
	w.Write([]byte(fmt.Sprintf(form, s)))
}

func NotImplemented(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
