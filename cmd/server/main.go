package main

import (
	"context"
	"github.com/fortuna91/go_ya_praktikum/internal/db"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fortuna91/go_ya_praktikum/internal/configs"
	"github.com/fortuna91/go_ya_praktikum/internal/fsstorage"
	"github.com/fortuna91/go_ya_praktikum/internal/handlers"
	"github.com/fortuna91/go_ya_praktikum/internal/middleware"
	"github.com/fortuna91/go_ya_praktikum/internal/server"
)

func main() {
	config := configs.SetServerConfig()

	r := server.NewRouter()
	server := &http.Server{Addr: config.Address, Handler: middleware.GzipHandle(r)}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigChan
		handlers.Storage.StoreBatchMetrics(context.Background(), *handlers.Metrics.List())
		handlers.Storage.Close()

		ctx, serverStopCtx := context.WithTimeout(context.Background(), 10*time.Second)
		err := server.Shutdown(ctx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
		log.Println("Server was stopped correctly")
	}()

	handlers.HashKey = config.Key

	if len(config.DB) > 0 {
		handlers.StoreMetrics = true
		handlers.Storage = db.New(config.DB)
		handlers.Storage.Create(context.Background())
	} else if len(config.StoreFile) > 0 {
		handlers.StoreMetrics = true
		handlers.Storage = fsstorage.New(config.StoreFile)
		if config.StoreInterval > 0 {
			storeTicker := time.NewTicker(config.StoreInterval)
			go fsstorage.StoreMetricsTicker(&handlers.Storage, storeTicker, &handlers.Metrics)
		} else {
			handlers.StoreMetricImmediately = true
		}
	} else {
		log.Println("Do not store metrics")
	}

	if config.Restore {
		storedMetrics := handlers.Storage.Restore()
		handlers.Metrics.RestoreMetrics(storedMetrics)
	}

	log.Printf("Start server on %s", config.Address)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
