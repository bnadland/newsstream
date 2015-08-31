package main

import (
	"errors"
	"fmt"
	"github.com/caser/gophernews"
	"github.com/jzelinskie/geddit"
	"io/ioutil"
	"net/http"
	//	"strconv"
	//	"github.com/jinzhu/now"
	"github.com/PuerkitoBio/purell"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"time"
)

type NewsItem struct {
	Title       string
	Url         string
	Source      string
	Raw         string `sql:"type:text"`
	HasRaw      bool
	Error       bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt time.Time
}

func saveNewsItem(db *gorm.DB, newsItem NewsItem) error {
	defer func() {
		if r := recover(); r != nil {
			log.WithFields(log.Fields{"func": "saveNewsItem", "r": r}).Error("Recovered")
		}
	}()
	url, err := purell.NormalizeURLString(newsItem.Url, purell.FlagsSafe)
	if err != nil {
		log.WithFields(log.Fields{"func": "saveNewsItem", "err": err}).Error("Url normalization failed")
		return err
	}
	newsItem.Url = url

	zeroTime := time.Time{}
	if newsItem.PublishedAt == zeroTime {
		newsItem.PublishedAt = time.Now()
	}

	if db.Where("url = ?", newsItem.Url).First(&NewsItem{}).RecordNotFound() {
		log.WithFields(log.Fields{
			"func":        "saveNewsItem",
			"Title":       newsItem.Title,
			"Url":         newsItem.Url,
			"PublishedAt": newsItem.PublishedAt,
		}).Info("New NewsItem")

		db.Save(&newsItem)
	}

	return nil
}

func crawlSubreddit(db *gorm.DB, subreddit string) error {
	r := geddit.NewSession("newsstream")

	submissions, err := r.SubredditSubmissions(subreddit, geddit.NewSubmissions, geddit.ListingOptions{
		Limit: 100})
	if err != nil {
		log.WithFields(log.Fields{"func": "main", "err": err, "subreddit": subreddit}).Error("Could not get submissions")
		return err
	}

	for _, submission := range submissions {
		if submission.IsSelf != true {
			saveNewsItem(db, NewsItem{
				Title:  submission.Title,
				Url:    submission.URL,
				Source: subreddit,
			})
		}
	}

	return nil
}

func crawlHackernews(db *gorm.DB) error {
	hn := gophernews.NewClient()
	if hn == nil {
		log.WithFields(log.Fields{"func": "crawlHackernews"}).Error("Could not init hackernews client")
		return errors.New("Could not init hackernews client")
	}

	submissionIds, err := hn.GetTop100()
	if err != nil {
		log.WithFields(log.Fields{"func": "crawlHackernews", "err": err}).Error("Could not get submissions")
		return err
	}

	for _, submissionId := range submissionIds {
		submission, err := hn.GetStory(submissionId)
		if err != nil {
			log.WithFields(log.Fields{"func": "crawlHackernews", "err": err, "submissionId": submissionId}).Error("Could not get submission")
			continue
		}
		saveNewsItem(db, NewsItem{
			Title:  submission.Title,
			Url:    submission.URL,
			Source: "hackernews",
		})
	}

	return nil
}

func getRawBodyFromUrl(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.WithFields(log.Fields{"func": "getRawBodyFromUrl", "url": url, "err": err}).Error("Could not fetch url")
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{"func": "getRawBodyFromUrl", "url": url, "err": err}).Error("Could read response")
		return "", err
	}

	return fmt.Sprintf("%s", body), nil
}

func crawlRawNewsItems(db *gorm.DB) error {
	for {
		newsItem := &NewsItem{}
		if db.Where("raw=''").Where("error = false").First(newsItem).RecordNotFound() {
			log.WithFields(log.Fields{"func": "crawlRawNewsItems"}).Info("No new NewsItems")
			time.Sleep(10 * time.Second)
			continue
		}
		body, err := getRawBodyFromUrl(newsItem.Url)
		if err != nil {
			log.WithFields(log.Fields{"func": "crawlRawNewsItems", "err": err, "url": newsItem.Url}).Info("Could not get raw body from url")
			newsItem.Error = true
			db.Save(newsItem)
			continue
		}

		log.WithFields(log.Fields{
			"func": "crawlRawNewsItem",
			"Url":  newsItem.Url,
		}).Info("Got raw body")

		newsItem.Raw = body
		newsItem.Error = false
		db.Save(newsItem)
		time.Sleep(1 * time.Second)
	}
}

func main() {
	db, err := gorm.Open("postgres", "postgres://newsstream:newsstream@127.0.0.1?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	//db.DropTable(&NewsItem{})
	//db.LogMode(true)

	db.AutoMigrate(&NewsItem{})

	/*
		for _, subreddit := range []string{"golang", "python", "worldnews"} {
			crawlSubreddit(&db, subreddit)
		}
	*/
	/*
		crawlHackernews(&db)
	*/
	crawlRawNewsItems(&db)
}
