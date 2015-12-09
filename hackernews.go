package main

import (
	"github.com/ChimeraCoder/gophernews"
	"github.com/Sirupsen/logrus"
	"time"
)

type HackernewsCrawler struct {
	interval time.Duration
	log      *logrus.Entry
	newItems chan<- Item
}

func NewHackernewsCrawler(newItems chan<- Item, interval time.Duration) *HackernewsCrawler {
	crawler := &HackernewsCrawler{
		interval: interval,
		log:      log.WithFields(logrus.Fields{"crawler": "hackernews"}),
		newItems: newItems,
	}
	return crawler
}

func (self *HackernewsCrawler) Crawl() {
	tick := time.Tick(self.interval)
	self.crawl()
	for _ = range tick {
		self.crawl()
	}
}

func (self *HackernewsCrawler) crawl() {
	self.log.Info("Crawling")
	defer self.log.Info("Crawling done")

	hn := gophernews.NewClient()

	ids, err := hn.GetTop100()
	if err != nil {
		self.log.Error("Unable to get submissions: ", err)
		return
	}

	for _, id := range ids {

		story, err := hn.GetStory(id)
		if err != nil {
			self.log.Errorf("Unable to get Story with id '%s': %s", id, err)
			return
		}

		self.newItems <- Item{
			Title:  story.Title,
			Url:    story.URL,
			Source: "hackernews",
		}
	}
}

func (self *HackernewsCrawler) Close() {
	self.log.Info("Shutting down")
}
