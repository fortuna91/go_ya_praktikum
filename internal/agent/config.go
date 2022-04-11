package agent

import "time"

type Config struct {
	Address        string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PolInterval    time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
}
