package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/fortuna91/go_ya_praktikum/internal/metrics"
)

func GetMetrics(count int64) []*metrics.Metric {
	var metricsList []*metrics.Metric
	var val float64
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	val = float64(mem.Alloc)
	metricsList = append(metricsList, &metrics.Metric{ID: "Alloc", MType: metrics.Gauge, Value: &val})
	val = float64(mem.BuckHashSys)
	metricsList = append(metricsList, &metrics.Metric{ID: "BuckHashSys", MType: metrics.Gauge, Value: &val})
	val = float64(mem.Frees)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "Frees", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: &mem.GCCPUFraction, ID: "GCCPUFraction", MType: metrics.Gauge})
	val = float64(mem.GCSys)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "GCSys", MType: metrics.Gauge})
	val = float64(mem.HeapAlloc)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "HeapAlloc", MType: metrics.Gauge})
	val = float64(mem.HeapIdle)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "HeapIdle", MType: metrics.Gauge})
	val = float64(mem.HeapInuse)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "HeapInuse", MType: metrics.Gauge})
	val = float64(mem.HeapObjects)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "HeapObjects", MType: metrics.Gauge})
	val = float64(mem.HeapReleased)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "HeapReleased", MType: metrics.Gauge})
	val = float64(mem.HeapSys)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "HeapSys", MType: metrics.Gauge})
	val = float64(mem.LastGC)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "LastGC", MType: metrics.Gauge})
	val = float64(mem.Lookups)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "Lookups", MType: metrics.Gauge})
	val = float64(mem.MCacheInuse)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "MCacheInuse", MType: metrics.Gauge})
	val = float64(mem.MCacheSys)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "MCacheSys", MType: metrics.Gauge})
	val = float64(mem.MSpanInuse)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "MSpanInuse", MType: metrics.Gauge})
	val = float64(mem.MSpanSys)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "MSpanSys", MType: metrics.Gauge})
	val = float64(mem.Mallocs)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "Mallocs", MType: metrics.Gauge})
	val = float64(mem.NextGC)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "NextGC", MType: metrics.Gauge})
	val = float64(mem.NumForcedGC)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "NumForcedGC", MType: metrics.Gauge})
	val = float64(mem.NumGC)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "NumGC", MType: metrics.Gauge})
	val = float64(mem.OtherSys)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "OtherSys", MType: metrics.Gauge})
	val = float64(mem.PauseTotalNs)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "PauseTotalNs", MType: metrics.Gauge})
	val = float64(mem.StackInuse)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "StackInuse", MType: metrics.Gauge})
	val = float64(mem.StackSys)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "StackSys", MType: metrics.Gauge})
	val = float64(mem.Sys)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "Sys", MType: metrics.Gauge})
	val = float64(mem.TotalAlloc)
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "TotalAlloc", MType: metrics.Gauge})

	metricsList = append(metricsList, &metrics.Metric{Delta: &count, ID: "PollCount", MType: metrics.Counter}) // TODO
	val = rand.Float64()
	metricsList = append(metricsList, &metrics.Metric{Value: &val, ID: "RandomValue", MType: metrics.Gauge})
	return metricsList
}

func SendRequest(client *http.Client, request *http.Request) int {
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	return response.StatusCode
}

func SendMetrics(metrics *[]*metrics.Metric) {
	client := http.Client{}
	for _, m := range *metrics {
		// request, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080/update/"+m.metricType+"/"+m.metricName+"/"+m.value, nil)
		body, err := json.Marshal(m)
		if err != nil {
			fmt.Printf("Cannot convert Metric to JSON: %v", err)
			continue
		}
		request, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080/update", bytes.NewReader(body))
		responseCode := SendRequest(&client, request)
		if responseCode != 200 {
			fmt.Printf("Error in request for %v: response code: %d", m.ID, responseCode)
		}
	}
}

func RunAgent() {
	fmt.Println("Start sending metrics...")

	ticker := time.NewTicker(2 * time.Second)
	var i int64 = 0
	for {
		<-ticker.C
		i++
		metrics := GetMetrics(i)
		if i%5 == 0 {
			SendMetrics(&metrics)
			// set i = 0 ??
		}
	}
}
