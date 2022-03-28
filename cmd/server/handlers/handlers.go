package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

/*var gaugeMetrics = []string{
	"Alloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
	"TotalAlloc",
	"PollCount",
	"RandomValue",
}*/

// var counterMetric = "PollCount"

var currCount int64 = 0

// var CountChannel = make(chan int64) maybe for feature

/*func Counter(c <-chan int64) {
	for v := range c {
		fmt.Println("here")
		currCount += v
	}
}*/

func Counter(c int64) {
	currCount += c
	fmt.Printf("Count is %d\n", currCount)
}

func contains(a []string, x string) bool {
	for _, v := range a {
		if x == v {
			return true
		}
	}
	return false
}

func SetGaugeMetric(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got it...")
	path := r.URL.Path
	items := strings.Split(path, "/")
	/*if !contains(gaugeMetrics, items[3]) {
		http.Error(w, "Unknown metric", http.StatusBadRequest)
		return
	}*/
	if len(items) < 5 {
		fmt.Printf("Not enough value. Path: %v\n", path)
		http.Error(w, "Not enough value", http.StatusNotFound)
		return
	}
	val := items[4]
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
	w.WriteHeader(http.StatusOK)
}

func SetCounterMetric(w http.ResponseWriter, r *http.Request) {
	// go Counter(CountChannel)
	path := r.URL.Path
	items := strings.Split(path, "/")
	/*if items[3] != counterMetric {
		http.Error(w, "Unknown metric", http.StatusBadRequest)
		return
	}*/
	// fixme: the same code
	if len(items) < 5 {
		fmt.Printf("Not enough value. Path: %v\n", path)
		http.Error(w, "Not enough value", http.StatusNotFound)
		return
	}
	val := items[4]
	if val == "" {
		http.Error(w, "Empty value", http.StatusNotFound)
		return
	}
	countVal, err := strconv.ParseInt(items[4], 10, 64)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		http.Error(w, "Wrong metric value", http.StatusBadRequest)
		return
	}
	// CountChannel <- countVal
	Counter(countVal)
	w.WriteHeader(http.StatusOK)
}
