package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"time"
)

type Newsstream struct {
	config    Config
	log       *logrus.Entry
	db        *gorm.DB
	newItems  chan Item
	fetchBody chan int
	crawlers  []Crawler
	done      bool
}

func NewNewsstream(config Config) (*Newsstream, error) {
	log.Info("Creating app")

	db, err := gorm.Open("postgres", config.Dsn)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	db.SetLogger(log.WithFields(logrus.Fields{"app": "gorm"}))
	db.AutoMigrate(&Item{})

	newsstream := &Newsstream{
		config:    config,
		log:       log.WithFields(logrus.Fields{"app": "newsstream"}),
		db:        &db,
		newItems:  make(chan Item),
		fetchBody: make(chan int),
	}

	// pull from config or db
	for _, subreddit := range []string{"golang", "programming", "python", "webdev"} {
		newsstream.addCrawler("reddit", subreddit)
	}
	newsstream.addCrawler("hackernews", "")

	return newsstream, nil
}

func (self *Newsstream) Run() {
	self.log.Info("Running")

	for i := 0; i < 5; i++ {
		go self.readability()
	}
	go self.handleItems()

	for _, crawler := range self.crawlers {
		go crawler.Crawl()
	}

	go self.webServer()

	for self.done == false {
		time.Sleep(1 * time.Second)
	}
}

func (self *Newsstream) Close() {
	self.log.Info("Shutting down")
	close(self.newItems)
	close(self.fetchBody)
	if self.db != nil {
		self.db.Close()
	}
	self.done = true
}
