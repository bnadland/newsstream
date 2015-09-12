package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/caser/gophernews"
	"github.com/jzelinskie/geddit"
)

type NewsItem struct {
	Id        string
	Title     string
	Url       string
	CreatedAt int
	Source    string
}

func crawlHackernews(db *bolt.DB) {
	source := "hackernews"

	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(source))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error: Cannot create bucket 'hackernews': %s\n", err)
	}

	client := gophernews.NewClient()
	storyIds, err := client.GetTop100()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, storyId := range storyIds {
		storySeen := false
		_ = db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(source))
			seen := b.Get([]byte(fmt.Sprintf("%v", storyId)))
			if seen != nil {
				storySeen = true
			}
			return nil
		})

		if storySeen {
			continue
		}

		story, err := client.GetStory(storyId)
		if err != nil {
			fmt.Printf("Error: Cannot get story with id %v", storyId)
			continue
		}

		newsItem := NewsItem{
			Id:        fmt.Sprintf("%v", story.ID),
			Title:     story.Title,
			Url:       story.URL,
			CreatedAt: story.Time,
			Source:    source,
		}

		err = db.Update(func(tx *bolt.Tx) error {
			sourceBucket := tx.Bucket([]byte(source))
			sourceItem, err := json.Marshal(story)
			if err != nil {
				return err
			}
			sourceBucket.Put([]byte(fmt.Sprintf("%v", storyId)), sourceItem)

			newsItemBucket := tx.Bucket([]byte("newsitem"))
			newsItemJson, err := json.Marshal(newsItem)
			if err != nil {
				return err
			}
			newsItemBucket.Put([]byte(fmt.Sprintf("%v-%v", newsItem.CreatedAt, newsItem.Id)), newsItemJson)

			return nil
		})
		if err != nil {
			fmt.Printf("Error: Could not save newsitem %s-%v\n", source, story.ID)
		}

		fmt.Printf("[%s] %s\n%s\n\n", source, newsItem.Title, newsItem.Url)
	}
}

func crawlSubreddit(db *bolt.DB, source string) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(source))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error: Cannot create bucket '%s': %s\n", source, err)
	}

	client := geddit.NewSession("newsstream by /u/bnadland")
	stories, err := client.SubredditSubmissions(source, geddit.NewSubmissions, geddit.ListingOptions{
		Limit: 100,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, story := range stories {
		storySeen := false
		_ = db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(source))
			seen := b.Get([]byte(story.ID))
			if seen != nil {
				storySeen = true
			}
			return nil
		})

		if storySeen {
			continue
		}
		newsItem := NewsItem{
			Id:        story.ID,
			Title:     story.Title,
			Url:       story.URL,
			CreatedAt: int(story.DateCreated),
			Source:    source,
		}

		err = db.Update(func(tx *bolt.Tx) error {
			sourceBucket := tx.Bucket([]byte(source))
			sourceItem, err := json.Marshal(story)
			if err != nil {
				return err
			}
			sourceBucket.Put([]byte(fmt.Sprintf("%v", story.ID)), sourceItem)

			newsItemBucket := tx.Bucket([]byte("newsitem"))
			newsItemJson, err := json.Marshal(newsItem)
			if err != nil {
				return err
			}
			newsItemBucket.Put([]byte(fmt.Sprintf("%v-%v", newsItem.CreatedAt, newsItem.Id)), newsItemJson)

			return nil
		})
		if err != nil {
			fmt.Printf("Error: Could not save newsitem %s-%v\n", source, story.ID)
		}

		fmt.Printf("[%s] %s\n%s\n\n", source, newsItem.Title, newsItem.Url)
	}
}

func main() {
	db, err := bolt.Open("newsstream.db", 0600, nil)
	if err != nil {
		fmt.Printf("Error: Cannot access 'newsstream.db': %s\n", err)
		return
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("newsitem"))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error: Cannot create bucket 'newsitem': %s\n", err)
	}

	crawlSubreddit(db, "golang")
	crawlSubreddit(db, "python")
	crawlSubreddit(db, "webdev")
	crawlHackernews(db)
}
