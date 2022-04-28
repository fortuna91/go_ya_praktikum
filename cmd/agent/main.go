package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fortuna91/go_ya_praktikum/internal/agent"
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
			log.Print("Signal interrupt triggered.")
			os.Exit(0)
		case syscall.SIGTERM:
			log.Print("Signal terminte triggered.")
			os.Exit(0)
		case syscall.SIGQUIT:
			log.Print("Signal quit triggered.")
			os.Exit(0)
		default:
			fmt.Println("Unknown signal.")
			os.Exit(1)
		}
	}()

	agent.RunAgent()
}
