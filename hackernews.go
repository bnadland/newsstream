package main

import (
	"github.com/caser/gophernews"
	"log"
	"time"
)

type HackernewsCrawler struct {
}

func (crawler *HackernewsCrawler) Crawl(items chan<- *Item) error {
	log.Printf("[HackernewsCrawler]\n")
	client := gophernews.NewClient()

	storyIds, err := client.GetTop100()
	if err != nil {
		return err
	}

	for _, storyId := range storyIds {
		story, err := client.GetStory(storyId)
		if err != nil {
			continue
		}

		items <- &Item{
			crawlerType: "hackernews",
			title:       story.Title,
			link:        story.URL,
			date:        time.Unix(int64(story.Time), 0),
		}
	}

	return nil
}
