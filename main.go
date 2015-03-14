package main

import (
	"github.com/bnadland/newsstream/item"
	"github.com/bnadland/newsstream/source"
)

func main() {
	items := make(chan item.Item)
	sources := make(chan source.Source)

	// feeds
	go source.Get(sources, "feed", "./feedurls")

	// hn
	go func(sources chan source.Source) {
		sources <- source.Source{
			Crawler: "hackernews",
			Name:    "news.ycombinator.com",
		}
	}(sources)

	// subreddits
	go source.Get(sources, "reddit", "./subreddits")

	// processing sources
	for i := 0; i < 3; i++ {
		go source.Process(sources, items)
	}

	// processing items
	item.Process(items)
}
