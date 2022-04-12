package main

import (
	"context"
	"fmt"
	"github.com/fortuna91/go_ya_praktikum/internal/configs"
	"github.com/fortuna91/go_ya_praktikum/internal/handlers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fortuna91/go_ya_praktikum/internal/run"
)

func main() {
	config := configs.ReadServerConfig()
	handlers.StoreFile = config.StoreFile

	r := run.NewRouter()
	server := &http.Server{Addr: config.Address, Handler: r}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigChan
		handlers.StoreMetrics(config.StoreFile)

		ctx, serverStopCtx := context.WithTimeout(context.Background(), 10*time.Second)
		err := server.Shutdown(ctx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
		log.Println("Server was stopped correctly")
	}()

	if config.Restore {
		handlers.Restore(config)
	}

	if config.StoreInterval > 0 {
		handlers.StoreMetricImmediately = false
		storeTicker := time.NewTicker(config.StoreInterval)
		go handlers.StoreMetricsTicker(storeTicker, config)
	} else {
		handlers.StoreMetricImmediately = true
	}

	fmt.Println("Start server")
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
