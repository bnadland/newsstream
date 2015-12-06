package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/jzelinskie/geddit"
	"time"
)

type RedditCrawler struct {
	subreddit string
	interval  time.Duration
	log       *logrus.Entry
	newItems  chan<- Item
}

func NewRedditCrawler(newItems chan<- Item, subreddit string, interval time.Duration) *RedditCrawler {
	crawler := &RedditCrawler{
		subreddit: subreddit,
		interval:  interval,
		log:       log.WithFields(logrus.Fields{"crawler": "reddit", "source": subreddit}),
		newItems:  newItems,
	}
	return crawler
}

func (self *RedditCrawler) Crawl() {
	tick := time.Tick(self.interval)
	self.crawl()
	for _ = range tick {
		self.crawl()
	}
}

func (self *RedditCrawler) crawl() {
	self.log.Info("Crawling")
	defer self.log.Info("Crawling done")

	r := geddit.NewSession("newsstream v0.1 by /u/bnadland")

	submissions, err := r.SubredditSubmissions(self.subreddit, geddit.DefaultPopularity, geddit.ListingOptions{Count: 100})
	if err != nil {
		self.log.Error("Unable to get submissions: ", err)
		return
	}

	for _, submission := range submissions {
		if submission.IsSelf {
			continue
		}
		self.newItems <- Item{
			Title:  submission.Title,
			Url:    submission.URL,
			Source: self.subreddit,
		}
	}
}

func (self *RedditCrawler) Close() {
	self.log.Info("Shutting down")
}
