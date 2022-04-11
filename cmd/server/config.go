package main

type Config struct {
	Address string `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
}
