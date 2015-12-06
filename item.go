package main

import (
	"github.com/PuerkitoBio/purell"
	"github.com/Sirupsen/logrus"
	"time"
)

type Item struct {
	Id          int       `json:"id"`
	Title       string    `sql:"type:text" json:"title"`
	Url         string    `sql:"type:text;unique" json:"url"`
	Body        string    `sql:"type:text" json:"omit"`
	Readability string    `sql:"type:text" json:"content"`
	Source      string    `json:"source"`
	CreatedAt   time.Time `json:"created"`
	UpdatedAt   time.Time `json:"updated"`
}

func (self *Newsstream) handleItem(item Item) {
	url, err := purell.NormalizeURLString(item.Url, purell.FlagsSafe)
	if err != nil {
		self.log.Warning("Unable to normalize url", url, err)
	} else {
		item.Url = url
	}
	self.db.Where(Item{Url: item.Url}).Attrs(item).FirstOrInit(&item)
	isNew := self.db.NewRecord(item)
	err = self.db.Save(&item).Error
	if err != nil {
		self.log.Error(err)
		return
	}
	if isNew == false {
		return
	}
	self.log.WithFields(logrus.Fields{"source": item.Source}).Printf("Item: %s", item.Title)
	self.fetchBody <- item.Id
}

func (self *Newsstream) handleItems() {
	for item := range self.newItems {
		self.handleItem(item)
	}
}
