package configs

import (
	"github.com/caarlos0/env/v6"
	"log"
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

func ReadAgentConfig() AgentConfig {
	var config AgentConfig
	err := env.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

func ReadServerConfig() ServerConfig {
	var config ServerConfig
	err := env.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}
