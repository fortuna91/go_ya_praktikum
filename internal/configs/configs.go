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
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	Key            string        `env:"KEY"`
}

type ServerConfig struct {
	Address       string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"30s"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore       bool          `env:"RESTORE" envDefault:"true"`
	Key           string        `env:"KEY"`
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
		flag.DurationVar(&config.ReportInterval, "r", 10*time.Second, "Report interval")
	}
	if _, ok := os.LookupEnv("POLL_INTERVAL"); !ok {
		flag.DurationVar(&config.PollInterval, "p", 2*time.Second, "Poll interval")
	}
	if _, ok := os.LookupEnv("KEY"); !ok {
		flag.StringVar(&config.Key, "k", "", "Key for hashing")
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
	addr := flag.String("a", "127.0.0.1:8080", "Address")
	interval := flag.Duration("i", 30*time.Second, "Store interval")
	file := flag.String("f", "/tmp/devops-metrics-db.json", "Store file name")
	restore := flag.Bool("r", true, "Restore")
	key := flag.String("k", "", "Key for ha	shing")
	flag.Parse()

	if _, ok := os.LookupEnv("ADDRESS"); !ok {
		config.Address = *addr
	}
	if _, ok := os.LookupEnv("STORE_INTERVAL"); !ok {
		config.StoreInterval = *interval
	}
	if _, ok := os.LookupEnv("STORE_FILE"); !ok {
		config.StoreFile = *file
	}
	if _, ok := os.LookupEnv("RESTORE"); !ok {
		config.Restore = *restore
	}
	if _, ok := os.LookupEnv("KEY"); !ok {
		config.Key = *key
	}
	return config
}
