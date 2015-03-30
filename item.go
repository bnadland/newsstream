package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/mauidude/go-readability"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type Context struct {
	item *Item
	save chan<- *Item
}

type Item struct {
	title        string
	link         string
	readability  string
	raw          string
	crawlerType  string
	crawlerName  string
	date         time.Time
	createdAt    time.Time
	saveRetries  int
	fetchRetries int
}

type stateFn func(*Context) stateFn

func fetch(ctx *Context) stateFn {
	log.Printf("[fetch] %s", ctx.item.link)

	r, err := http.Get(ctx.item.link)
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

	ctx.item.raw = string(body)

	return extractArticle
}

func fetchLater(ctx *Context) stateFn {
	log.Printf("[fetchLater] %s", ctx.item.link)
	ctx.item.fetchRetries += 1
	switch ctx.item.fetchRetries {
	case 1, 2, 3:
		time.Sleep(10 * time.Second)
		return fetch
	default:
		log.Printf("Error: too many retries")
		return save
	}
}

func extractArticle(ctx *Context) stateFn {
	log.Printf("[extractArticle] %s", ctx.item.link)
	doc, err := readability.NewDocument(ctx.item.raw)
	if err != nil {
		log.Printf("Error: %s", err)
		return save
	}

	text, err := goquery.NewDocumentFromReader(strings.NewReader(doc.Content()))
	if err != nil {
		log.Printf("Error: %s", err)
		return save
	}

	ctx.item.readability = text.Text()

	return trimWhitespace
}

func trimWhitespace(ctx *Context) stateFn {
	log.Printf("[extractArticle] %s", ctx.item.link)
	ctx.item.readability = strings.TrimSpace(ctx.item.readability)
	return save
}

func save(ctx *Context) stateFn {
	log.Printf("[save] %s", ctx.item.link)
	ctx.save <- ctx.item
	return nil
}

func process(ctx *Context) {
	log.Printf("[process] %+v", ctx.item)
	for state := fetch; state != nil; {
		state = state(ctx)
	}
}

func processItems(save chan<- *Item, items <-chan *Item) {
	log.Printf("[processItems] starting worker")
	for {
		item := <-items
		ctx := &Context{
			item: item,
			save: save,
		}
		process(ctx)
	}
}
