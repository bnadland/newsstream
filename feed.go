package main

import (
	"github.com/nmeum/go-feedparser"
	"github.com/nmeum/go-feedparser/atom"
	"github.com/nmeum/go-feedparser/rss"
	"log"
	"net/http"
)

type FeedCrawler struct {
	name string
}

func (crawler *FeedCrawler) Crawl(items chan<- *Item) error {
	log.Printf("[FeedCrawler] %s\n", crawler.name)

	response, err := http.Get(crawler.name)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	parsers := []feedparser.FeedFunc{rss.Parse, atom.Parse}
	feed, err := feedparser.Parse(response.Body, parsers)
	if err != nil {
		return err
	}

	for _, feeditem := range feed.Items {
		items <- &Item{
			crawlerType: "feed",
			crawlerName: crawler.name,
			title:       feeditem.Title,
			link:        feeditem.Link,
			date:        feeditem.Date,
		}
	}

	return nil
}
