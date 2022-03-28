package run

import (
	"fmt"
	"github.com/fortuna91/go_ya_praktikum/cmd/server/handlers"
	"net/http"
)

func NewServer() {
	http.HandleFunc("/update/gauge/", handlers.SetGaugeMetric)
	http.HandleFunc("/update/counter/", handlers.SetCounterMetric)
	http.HandleFunc("/update/", handlers.NotImplemented)

	server := &http.Server{
		Addr: "127.0.0.1:8080",
	}
	fmt.Println("Start server")
	server.ListenAndServe()
}
