package item

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/boltdb/bolt"
	"github.com/mauidude/go-readability"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type Item struct {
	Title        string
	Link         string
	Content      string
	Raw          string
	Source       string
	Date         time.Time
	CreatedAt    time.Time
	SaveRetries  int
	FetchRetries int
}

type stateFn func(*Item) stateFn

func fetch(item *Item) stateFn {
	log.Printf("[fetch] %s", item.Link)

	r, err := http.Get(item.Link)
	if err != nil {
		log.Printf("Error: %s", err)
		return fetchLater
	}
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		log.Printf("Error: %s", err)
		return fetchLater
	}

	item.Raw = string(body)
	return extractArticle
}

func fetchLater(item *Item) stateFn {
	log.Printf("[fetchLater] %s", item.Link)
	item.FetchRetries += 1
	switch item.FetchRetries {
	case 1, 2, 3:
		time.Sleep(10 * time.Second)
		return fetch
	case 4, 5, 6, 7:
		time.Sleep(1 * time.Minute)
		return fetch
	default:
		log.Printf("Error: too many retries")
		return save
	}
}

func extractArticle(item *Item) stateFn {
	log.Printf("[extractArticle] %s", item.Link)
	doc, err := readability.NewDocument(item.Raw)
	if err != nil {
		log.Printf("Error: %s", err)
		return save
	}

	text, err := goquery.NewDocumentFromReader(strings.NewReader(doc.Content()))
	if err != nil {
		log.Printf("Error: %s", err)
		return save
	}

	item.Content = text.Text()

	return save
}

func save(item *Item) stateFn {
	log.Printf("[save] %s", item.Link)

	db, err := bolt.Open("newsstream.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Printf("Error: %s", err)
		return saveLater
	}
	defer db.Close()

	jsonDoc, err := json.Marshal(item)
	if err != nil {
		log.Printf("Error: %s", err)
		return saveLater
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("items"))
		if err != nil {
			return err
		}
		err = b.Put([]byte(item.Link), jsonDoc)
		return err
	})

	if err != nil {
		log.Printf("Error: %s", err)
		return saveLater
	}

	return nil
}

func saveLater(item *Item) stateFn {
	log.Printf("[saveLater] %s", item.Link)
	item.SaveRetries += 1
	switch item.SaveRetries {
	case 1, 2, 3:
		time.Sleep(10 * time.Second)
		return save
	case 4, 5, 6, 7:
		time.Sleep(1 * time.Minute)
		return save
	default:
		log.Printf("Error: too many retries")
		return nil
	}

}

func isNew(item *Item) stateFn {
	log.Printf("[isNew] %s", item.Link)

	db, err := bolt.Open("newsstream.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Printf("Error: %s", err)
		return saveLater
	}
	defer db.Close()

	tx, err := db.Begin(false)
	if err != nil {
		log.Printf("Error: %s", err)
		return fetch
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte("items"))
	if b == nil {
		// if the bucket does not exist the item is not in the db
		return fetch
	}

	doc := b.Get([]byte(item.Link))
	if doc != nil {
		// item is already in the db
		return nil
	}

	return fetch
}

func process(item Item) {
	log.Printf("[process] %+v", item)
	for state := isNew; state != nil; {
		state = state(&item)
	}
}

func Process(items <-chan Item) {
	count := 0
	for {
		if count > 10 {
			break
		}
		item := <-items
		process(item)
		count += 1
	}
}
