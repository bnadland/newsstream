package main

import (
	"github.com/jzelinskie/geddit"
	"log"
	"time"
)

type RedditCrawler struct {
	name string
}

func (crawler *RedditCrawler) Crawl(items chan *Item) error {
	log.Printf("[RedditCrawler] %s\n", crawler.name)
	r := geddit.NewSession("newsstream by /u/bnadland")
	submissions, err := r.SubredditSubmissions(crawler.name)
	if err != nil {
		return err
	}
	for _, submission := range submissions {
		item := &Item{
			title:       submission.Title,
			link:        submission.URL,
			date:        time.Unix(int64(submission.DateCreated), 0),
			crawlerType: "reddit",
			crawlerName: crawler.name,
		}
		items <- item
	}
	return nil
}
