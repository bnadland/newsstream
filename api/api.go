package api

import (
	"flag"
	"github.com/blevesearch/bleve"
	"log"
	"net/http"
)

type Api struct {
	Dsn    string
	Port   string
	Events chan Event
	Search *bleve.Index
}

type Event struct {
	Type  string
	Value string
}

var App Api

func Init() {
	flag.StringVar(&App.Dsn, "dsn", "newsstream.db", "database filename")
	flag.StringVar(&App.Port, "port", ":8080", "http port, e.g. 127.0.0.1:8080")
	flag.Parse()
	App.Events = make(chan Event)
}

func (app *Api) Run() {
	go func() {
		for {
			event := <-app.Events
			log.Printf("[Event] Type=%s Value=%s", event.Type, event.Value)
		}
	}()
	http.HandleFunc("/api/items", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello, world"))
	})
	http.ListenAndServe(app.Port, nil)
}
