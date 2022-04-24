package main

import (
	"context"
	"fmt"
	"github.com/fortuna91/go_ya_praktikum/internal/configs"
	"github.com/fortuna91/go_ya_praktikum/internal/handlers"
	"github.com/fortuna91/go_ya_praktikum/internal/middleware"
	"github.com/fortuna91/go_ya_praktikum/internal/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fortuna91/go_ya_praktikum/internal/server"
)

func main() {
	config := configs.SetServerConfig()
	handlers.StoreFile = config.StoreFile

	r := server.NewRouter()
	server := &http.Server{Addr: config.Address, Handler: middleware.GzipHandle(r)}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigChan
		storage.StoreMetrics(&handlers.Metrics, config.StoreFile)

		ctx, serverStopCtx := context.WithTimeout(context.Background(), 10*time.Second)
		err := server.Shutdown(ctx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
		log.Println("Server was stopped correctly")
	}()

	if config.Restore {
		storage.Restore(&handlers.Metrics, config)
	}

	// true by default
	if config.StoreInterval > 0 {
		handlers.StoreMetricImmediately = false
		storeTicker := time.NewTicker(config.StoreInterval)
		go storage.StoreMetricsTicker(storeTicker, &handlers.Metrics, config)
	}

	handlers.HashKey = config.Key
	handlers.DbAddress = config.DB

	fmt.Println("Start server on", config.Address)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
