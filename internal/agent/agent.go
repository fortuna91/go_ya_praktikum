package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/fortuna91/go_ya_praktikum/internal/configs"
	"github.com/fortuna91/go_ya_praktikum/internal/metrics"
)

func getFloat64Pointer(val uint64) *float64 {
	fVal := float64(val)
	return &fVal
}

func GetMetrics(count int64) []*metrics.Metric {
	var metricsList []*metrics.Metric
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	metricsList = append(metricsList, &metrics.Metric{ID: "Alloc", MType: metrics.Gauge, Value: getFloat64Pointer(mem.Alloc)})
	metricsList = append(metricsList, &metrics.Metric{ID: "BuckHashSys", MType: metrics.Gauge, Value: getFloat64Pointer(mem.BuckHashSys)})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.Frees), ID: "Frees", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: &mem.GCCPUFraction, ID: "GCCPUFraction", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.GCSys), ID: "GCSys", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.HeapAlloc), ID: "HeapAlloc", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.HeapIdle), ID: "HeapIdle", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.HeapInuse), ID: "HeapInuse", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.HeapObjects), ID: "HeapObjects", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.HeapReleased), ID: "HeapReleased", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.HeapSys), ID: "HeapSys", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.LastGC), ID: "LastGC", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.Lookups), ID: "Lookups", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.MCacheInuse), ID: "MCacheInuse", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.MCacheSys), ID: "MCacheSys", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.MSpanInuse), ID: "MSpanInuse", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.MSpanSys), ID: "MSpanSys", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.Mallocs), ID: "Mallocs", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.NextGC), ID: "NextGC", MType: metrics.Gauge})
	numForced := float64(mem.NumForcedGC)
	metricsList = append(metricsList, &metrics.Metric{Value: &numForced, ID: "NumForcedGC", MType: metrics.Gauge})
	numGC := float64(mem.NumGC)
	metricsList = append(metricsList, &metrics.Metric{Value: &numGC, ID: "NumGC", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.OtherSys), ID: "OtherSys", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.PauseTotalNs), ID: "PauseTotalNs", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.StackInuse), ID: "StackInuse", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.StackSys), ID: "StackSys", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.Sys), ID: "Sys", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: getFloat64Pointer(mem.TotalAlloc), ID: "TotalAlloc", MType: metrics.Gauge})

	metricsList = append(metricsList, &metrics.Metric{Delta: &count, ID: "PollCount", MType: metrics.Counter}) // TODO
	val25 := rand.Float64()
	metricsList = append(metricsList, &metrics.Metric{Value: &val25, ID: "RandomValue", MType: metrics.Gauge})
	return metricsList
}

func SendRequest(client *http.Client, request *http.Request) int {
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer response.Body.Close()
	return response.StatusCode
}

func SendMetrics(metricsList *[]*metrics.Metric, config configs.AgentConfig) {
	client := http.Client{}
	if len(config.Key) > 0 {
		for _, m := range *metricsList {
			m.SetHash(config.Key)
		}
	}
	for _, m := range *metricsList {
		body, err := json.Marshal(m)
		if err != nil {
			fmt.Printf("Cannot convert Metric to JSON: %v", err)
			continue
		}
		request, _ := http.NewRequest(http.MethodPost, "http://"+config.Address+"/update", bytes.NewReader(body))
		responseCode := SendRequest(&client, request)
		if responseCode != 200 {
			fmt.Printf("Error in request for %v: response code: %d", m.ID, responseCode)
		}
	}
}

func RunAgent() {
	fmt.Println("Start sending metrics...")

	config := configs.SetAgentConfig()

	pollTicker := time.NewTicker(config.PollInterval)
	reportTicker := time.NewTicker(config.ReportInterval)
	ch := make(chan []*metrics.Metric)

	go func() {
		for {
			<-ch
		}
	}()

	go func() {
		for {
			<-reportTicker.C
			metrics := <-ch
			fmt.Println("Send metrics...")
			SendMetrics(&metrics, config)
		}
	}()

	for {
		var i int64 = 0
		for {
			<-pollTicker.C
			i++
			metrics := GetMetrics(i)
			ch <- metrics
		}
	}
}
