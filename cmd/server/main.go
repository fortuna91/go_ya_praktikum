package main

import (
	"fmt"
	"github.com/fortuna91/go_ya_praktikum/cmd/server/run"
	"os"
	"os/signal"
	"syscall"
)

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

	run.NewServer()
}
