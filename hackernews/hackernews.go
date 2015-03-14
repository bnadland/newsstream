package hackernews

import (
	"github.com/bnadland/newsstream/item"
	"github.com/caser/gophernews"
	"time"
)

func Get(items chan<- item.Item) error {
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

		items <- item.Item{
			Source: "news.ycombinator.com",
			Title:  story.Title,
			Link:   story.URL,
			Date:   time.Unix(int64(story.Time), 0),
		}
	}

	return nil
}
