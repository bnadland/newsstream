package main

import (
	"time"
)

type Crawler interface {
	Crawl()
	Close()
}

func (self *Newsstream) addCrawler(crawlerType string, source string) {
	self.log.Printf("Adding crawler '%s' for '%s'", crawlerType, source)

	var crawler Crawler

	switch crawlerType {
	case "reddit":
		crawler = NewRedditCrawler(self.newItems, source, 1*time.Hour)
	default:
		self.log.Error("Unknown crawler type '%s'", crawlerType)
	}

	if crawler != nil {
		self.crawlers = append(self.crawlers, crawler)
	}
}
