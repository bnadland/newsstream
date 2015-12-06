package main

import (
	"github.com/Sirupsen/logrus"
	"os"
)

var log = logrus.New()

type Config struct {
	Listen string
	Dsn    string
}

func NewConfig() Config {
	config := Config{}

	config.Listen = os.Getenv("APP_LISTEN")
	if config.Listen == "" {
		config.Listen = "127.0.0.1:8080"
	}

	config.Dsn = os.Getenv("APP_DSN")
	if config.Dsn == "" {
		config.Dsn = "user=newsstream password=newsstream dbname=newsstream sslmode=disable"
	}

	setupLogger()

	os.Clearenv()
	return config
}

func setupLogger() {
	log.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
	}
}
