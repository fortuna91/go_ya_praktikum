package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

type Metric struct {
	value      string
	metricName string
	metricType string
}

func GetMetrics(count int64) []Metric {
	var metrics []Metric
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	metrics = append(metrics, Metric{strconv.FormatUint(mem.Alloc, 10), "Alloc", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.BuckHashSys, 10), "BuckHashSys", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.Frees, 10), "Frees", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatFloat(mem.GCCPUFraction, 'f', 2, 64), "GCCPUFraction", "gauge"}) // fixme
	metrics = append(metrics, Metric{strconv.FormatUint(mem.GCSys, 10), "GCSys", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.HeapAlloc, 10), "HeapAlloc", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.HeapIdle, 10), "HeapIdle", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.HeapInuse, 10), "HeapInuse", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.HeapObjects, 10), "HeapObjects", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.HeapReleased, 10), "HeapReleased", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.HeapSys, 10), "HeapSys", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.LastGC, 10), "LastGC", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.Lookups, 10), "Lookups", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.MCacheInuse, 10), "MCacheInuse", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.MCacheSys, 10), "MCacheSys", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.MSpanInuse, 10), "MSpanInuse", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.MSpanSys, 10), "MSpanSys", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.Mallocs, 10), "Mallocs", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.NextGC, 10), "NextGC", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(uint64(mem.NumForcedGC), 10), "NumForcedGC", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(uint64(mem.NumGC), 10), "NumGC", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.OtherSys, 10), "OtherSys", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.PauseTotalNs, 10), "PauseTotalNs", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.StackInuse, 10), "StackInuse", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.StackSys, 10), "StackSys", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.Sys, 10), "Sys", "gauge"})
	metrics = append(metrics, Metric{strconv.FormatUint(mem.TotalAlloc, 10), "TotalAlloc", "gauge"})

	metrics = append(metrics, Metric{strconv.FormatInt(count, 10), "PollCount", "counter"}) // TODO
	metrics = append(metrics, Metric{strconv.FormatFloat(rand.Float64(), 'f', 2, 64), "RandomValue", "gauge"})
	return metrics
}

func SendRequest(client *http.Client, request *http.Request) int {
	request.Header.Set("Content-Type", "text/plain")
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	return response.StatusCode
}

func SendMetrics(metrics *[]Metric) {
	client := http.Client{}
	for _, m := range *metrics {
		request, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080/update/"+m.metricType+"/"+m.metricName+"/"+m.value, nil)
		responseCode := SendRequest(&client, request)
		if responseCode != 200 {
			fmt.Printf("Error in request %v/%v: response code: %d", m.metricType, m.metricName, responseCode)
		}
	}
}

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		s := <-sigChan
		switch s {
		case syscall.SIGINT:
			fmt.Println("Signal interrupt triggered.")
			os.Exit(0)
		case syscall.SIGTERM:
			fmt.Println("Signal terminte triggered.")
			os.Exit(0)
		case syscall.SIGQUIT:
			fmt.Println("Signal quit triggered.")
			os.Exit(0)
		default:
			fmt.Println("Unknown signal.")
			os.Exit(1)
		}
	}()

	fmt.Println("Start sending metrics...")
	var i int64 = 0
	for true {
		i++
		metrics := GetMetrics(i)
		time.Sleep(time.Second * 2)
		if i%5 == 0 {
			SendMetrics(&metrics)
			// set i = 0 ??
		}
	}
}
