package main

import (
	"context"
	"github.com/fortuna91/go_ya_praktikum/internal/db"
	"github.com/fortuna91/go_ya_praktikum/internal/metrics"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fortuna91/go_ya_praktikum/internal/configs"
	"github.com/fortuna91/go_ya_praktikum/internal/handlers"
	"github.com/fortuna91/go_ya_praktikum/internal/middleware"
	"github.com/fortuna91/go_ya_praktikum/internal/server"
	"github.com/fortuna91/go_ya_praktikum/internal/storage"
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
		if handlers.UseDB {
			storage.StoreMetrics(&handlers.Metrics, config.StoreFile)
			handlers.DB.Close()
		}

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
		handlers.DB = db.Connect(config.DB)
		handlers.DBAddress = config.DB
		// defer handlers.DB.Close()
		handlers.UseDB = true
		db.CreateTable(handlers.DB)
	} else if config.StoreInterval > 0 {
		// true by default
		handlers.StoreMetricImmediately = false
		storeTicker := time.NewTicker(config.StoreInterval)
		go storage.StoreMetricsTicker(storeTicker, &handlers.Metrics, config)
	}

	if config.Restore {
		var storedMetrics map[string]*metrics.Metric
		if len(config.DB) > 0 {
			storedMetrics = db.Restore(handlers.DB)
		} else {
			storedMetrics = storage.Restore(config.StoreFile)
		}
		handlers.Metrics.RestoreMetrics(storedMetrics)
	}

	log.Printf("Start server on %s", config.Address)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
