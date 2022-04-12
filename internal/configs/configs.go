package configs

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
	"os"
	"time"
)

type AgentConfig struct {
	Address        string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PolInterval    time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
}

type ServerConfig struct {
	Address       string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"30s"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore       bool          `env:"RESTORE" envDefault:"true"`
}

func SetAgentConfig() AgentConfig {
	var config AgentConfig
	err := env.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}
	if _, ok := os.LookupEnv("ADDRESS"); !ok {
		flag.StringVar(&config.Address, "a", "127.0.0.1:8080", "Address")
	}
	if _, ok := os.LookupEnv("REPORT_INTERVAL"); !ok {
		flag.DurationVar(&config.ReportInterval, "i", 10*time.Second, "Report interval")
	}
	if _, ok := os.LookupEnv("POLL_INTERVAL"); !ok {
		flag.DurationVar(&config.PolInterval, "i", 2*time.Second, "Poll interval")
	}
	flag.Parse()
	return config
}

func SetServerConfig() ServerConfig {
	var config ServerConfig
	err := env.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}
	if _, ok := os.LookupEnv("ADDRESS"); !ok {
		flag.StringVar(&config.Address, "a", "127.0.0.1:8080", "Address")
	}
	if _, ok := os.LookupEnv("STORE_INTERVAL"); !ok {
		flag.DurationVar(&config.StoreInterval, "i", 30*time.Second, "Address")
	}
	if _, ok := os.LookupEnv("STORE_FILE"); !ok {
		flag.StringVar(&config.StoreFile, "f", "/tmp/devops-metrics-db.json", "Store file name")
	}
	if _, ok := os.LookupEnv("RESTORE"); !ok {
		flag.BoolVar(&config.Restore, "r", true, "Address")
	}
	flag.Parse()
	return config
}
