package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/fortuna91/go_ya_praktikum/internal/metrics"
)

func GetMetrics(count int64) []*metrics.Metric {
	var metricsList []*metrics.Metric
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	val := float64(mem.Alloc)
	metricsList = append(metricsList, &metrics.Metric{ID: "Alloc", MType: metrics.Gauge, Value: &val})
	val0 := float64(mem.BuckHashSys)
	metricsList = append(metricsList, &metrics.Metric{ID: "BuckHashSys", MType: metrics.Gauge, Value: &val0})
	val1 := float64(mem.Frees)
	metricsList = append(metricsList, &metrics.Metric{Value: &val1, ID: "Frees", MType: metrics.Gauge})
	metricsList = append(metricsList, &metrics.Metric{Value: &mem.GCCPUFraction, ID: "GCCPUFraction", MType: metrics.Gauge})
	val2 := float64(mem.GCSys)
	metricsList = append(metricsList, &metrics.Metric{Value: &val2, ID: "GCSys", MType: metrics.Gauge})
	val3 := float64(mem.HeapAlloc)
	metricsList = append(metricsList, &metrics.Metric{Value: &val3, ID: "HeapAlloc", MType: metrics.Gauge})
	val4 := float64(mem.HeapIdle)
	metricsList = append(metricsList, &metrics.Metric{Value: &val4, ID: "HeapIdle", MType: metrics.Gauge})
	val5 := float64(mem.HeapInuse)
	metricsList = append(metricsList, &metrics.Metric{Value: &val5, ID: "HeapInuse", MType: metrics.Gauge})
	val6 := float64(mem.HeapObjects)
	metricsList = append(metricsList, &metrics.Metric{Value: &val6, ID: "HeapObjects", MType: metrics.Gauge})
	val7 := float64(mem.HeapReleased)
	metricsList = append(metricsList, &metrics.Metric{Value: &val7, ID: "HeapReleased", MType: metrics.Gauge})
	val8 := float64(mem.HeapSys)
	metricsList = append(metricsList, &metrics.Metric{Value: &val8, ID: "HeapSys", MType: metrics.Gauge})
	val9 := float64(mem.LastGC)
	metricsList = append(metricsList, &metrics.Metric{Value: &val9, ID: "LastGC", MType: metrics.Gauge})
	val10 := float64(mem.Lookups)
	metricsList = append(metricsList, &metrics.Metric{Value: &val10, ID: "Lookups", MType: metrics.Gauge})
	val11 := float64(mem.MCacheInuse)
	metricsList = append(metricsList, &metrics.Metric{Value: &val11, ID: "MCacheInuse", MType: metrics.Gauge})
	val12 := float64(mem.MCacheSys)
	metricsList = append(metricsList, &metrics.Metric{Value: &val12, ID: "MCacheSys", MType: metrics.Gauge})
	val13 := float64(mem.MSpanInuse)
	metricsList = append(metricsList, &metrics.Metric{Value: &val13, ID: "MSpanInuse", MType: metrics.Gauge})
	val14 := float64(mem.MSpanSys)
	metricsList = append(metricsList, &metrics.Metric{Value: &val14, ID: "MSpanSys", MType: metrics.Gauge})
	val15 := float64(mem.Mallocs)
	metricsList = append(metricsList, &metrics.Metric{Value: &val15, ID: "Mallocs", MType: metrics.Gauge})
	val16 := float64(mem.NextGC)
	metricsList = append(metricsList, &metrics.Metric{Value: &val16, ID: "NextGC", MType: metrics.Gauge})
	val17 := float64(mem.NumForcedGC)
	metricsList = append(metricsList, &metrics.Metric{Value: &val17, ID: "NumForcedGC", MType: metrics.Gauge})
	val18 := float64(mem.NumGC)
	metricsList = append(metricsList, &metrics.Metric{Value: &val18, ID: "NumGC", MType: metrics.Gauge})
	val19 := float64(mem.OtherSys)
	metricsList = append(metricsList, &metrics.Metric{Value: &val19, ID: "OtherSys", MType: metrics.Gauge})
	val20 := float64(mem.PauseTotalNs)
	metricsList = append(metricsList, &metrics.Metric{Value: &val20, ID: "PauseTotalNs", MType: metrics.Gauge})
	val21 := float64(mem.StackInuse)
	metricsList = append(metricsList, &metrics.Metric{Value: &val21, ID: "StackInuse", MType: metrics.Gauge})
	val22 := float64(mem.StackSys)
	metricsList = append(metricsList, &metrics.Metric{Value: &val22, ID: "StackSys", MType: metrics.Gauge})
	val23 := float64(mem.Sys)
	metricsList = append(metricsList, &metrics.Metric{Value: &val23, ID: "Sys", MType: metrics.Gauge})
	val24 := float64(mem.TotalAlloc)
	metricsList = append(metricsList, &metrics.Metric{Value: &val24, ID: "TotalAlloc", MType: metrics.Gauge})

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
	}
	defer response.Body.Close()
	return response.StatusCode
}

func SendRequest2(client *http.Client, request *http.Request) (int, []byte) {
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	respBody, _ := ioutil.ReadAll(response.Body)
	return response.StatusCode, respBody
}

func SendMetrics(metricsList *[]*metrics.Metric) {
	client := http.Client{}
	for _, m := range *metricsList {
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
		/*m2 := metrics.Metric{ID: m.ID, MType: m.MType}
		b2, _ := json.Marshal(m2)
		r2, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080/value/", bytes.NewReader(b2))
		_, responseBody := SendRequest2(&client, r2)
		metricRequest := metrics.Metric{}
		json.Unmarshal(responseBody, &metricRequest)
		if m.MType == metrics.Gauge {
			fmt.Printf("%v -> %v\n", *m.Value, *metricRequest.Value)
		}*/
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
		if i%2 == 0 {
			fmt.Println("Send metrics...")
			SendMetrics(&metrics)
			// set i = 0 ??
		}
	}
}
