package feed

import (
	"github.com/bnadland/newsstream/item"
	"github.com/nmeum/go-feedparser"
	"github.com/nmeum/go-feedparser/atom"
	"github.com/nmeum/go-feedparser/rss"
	"net/http"
)

func Get(items chan<- item.Item, feedurl string) error {
	response, err := http.Get(feedurl)
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
		items <- item.Item{
			Source: feedurl,
			Title:  feeditem.Title,
			Link:   feeditem.Link,
			Date:   feeditem.Date,
		}
	}

	return nil
}
