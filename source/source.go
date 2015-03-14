package source

import (
	"bufio"
	"github.com/bnadland/newsstream/feed"
	"github.com/bnadland/newsstream/hackernews"
	"github.com/bnadland/newsstream/item"
	"github.com/bnadland/newsstream/reddit"
	"log"
	"os"
)

type Source struct {
	Crawler string
	Name    string
}

func Get(sources chan<- Source, crawlername string, filename string) error {
	names, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer names.Close()

	scanner := bufio.NewScanner(names)
	for scanner.Scan() {
		sources <- Source{
			Crawler: crawlername,
			Name:    scanner.Text(),
		}
	}
	return nil
}

func Process(sources <-chan Source, items chan<- item.Item) {
	for {
		source := <-sources
		log.Printf("processing %+v", source)
		switch source.Crawler {
		case "feed":
			feed.Get(items, source.Name)
		case "hackernews":
			hackernews.Get(items)
		case "reddit":
			reddit.Get(items, source.Name)
		default:
			log.Printf("Unknown source: %+v\n", source)
		}
	}
}
