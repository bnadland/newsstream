package main

import (
	"bufio"
	"encoding/json"
	"github.com/blevesearch/bleve"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

type Config struct {
	dsn  string
	port string
}

type Application struct {
	database *bolt.DB
	search   *bleve.Index
	config   Config
	items    chan *Item
	save     chan *Item
}

func NewApplication(config Config) (*Application, error) {
	app := &Application{}
	app.config.dsn = config.dsn
	db, err := bolt.Open(config.dsn, 0600, nil)
	if err != nil {
		return nil, err
	}
	app.database = db
	app.items = make(chan *Item)
	app.save = make(chan *Item)
	return app, nil
}

func (app *Application) Close() {
	app.database.Close()
}

func (app *Application) Run() {

	for worker := 0; worker < 3; worker++ {
		go processItems(app.save, app.items)
	}

	type Source struct {
		crawler string
		name    string
	}

	sources := make(chan *Source)

	for worker := 0; worker < 3; worker++ {
		go func(sources chan *Source, items chan *Item) {
			for {
				source := <-sources
				switch source.crawler {
				case "reddit":
					crawler := &RedditCrawler{
						name: source.name,
					}
					crawler.Crawl(items)
					break
				case "hackernews":
					crawler := &HackernewsCrawler{}
					crawler.Crawl(items)
					break
				case "feed":
					crawler := &FeedCrawler{
						name: source.name,
					}
					crawler.Crawl(items)
					break
				default:
					log.Printf("Unknown source: %+v", source)
				}
			}
		}(sources, app.items)
	}

	for _, name := range app.getLines("reddit.sources") {
		sources <- &Source{
			crawler: "reddit",
			name:    name,
		}
	}

	sources <- &Source{
		crawler: "hackernews",
	}

	for _, name := range app.getLines("feed.sources") {
		sources <- &Source{
			crawler: "feed",
			name:    name,
		}
	}

	app.saveItems()
}

func (app *Application) getLines(fileName string) []string {
	names, err := os.Open(fileName)
	if err != nil {
		log.Printf("[getLines: %s] %s", fileName, err)
		return []string{}
	}
	defer names.Close()

	lines := []string{}
	scanner := bufio.NewScanner(names)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func (app *Application) saveItems() {
	for {
		item := <-app.save
		jsonItem, err := json.Marshal(item)
		if err != nil {
			log.Println(err)
			continue
		}
		err = app.database.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("item"))
			if err != nil {
				return err
			}

			err = b.Put([]byte(item.link), jsonItem)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}
}

type Crawler interface {
	Crawl(items <-chan *Item) error
}
