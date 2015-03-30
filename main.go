package main

import (
	"flag"
)

func main() {
	config := Config{}
	flag.StringVar(&config.port, "port", ":8080", "hostname:port string to listen on")
	flag.StringVar(&config.dsn, "dsn", "./newsstream.db", "database path")
	flag.Parse()

	app, err := NewApplication(config)
	if err != nil {
		panic(err)
	}
	defer app.Close()
	app.Run()
}
