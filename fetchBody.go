package main

import (
	"github.com/mauidude/go-readability"
	"github.com/microcosm-cc/bluemonday"
	"io/ioutil"
	"net/http"
)

func (self *Newsstream) readability() {
	for itemId := range self.fetchBody {
		var item Item
		err := self.db.First(&item, itemId).Error
		if err != nil {
			self.log.Error(err)
			continue
		}

		r, err := http.Get(item.Url)
		if err != nil {
			self.log.Error(err)
			continue
		}
		if r.StatusCode > 300 {
			self.log.Error(err)
			continue
		}
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			self.log.Error(err)
			continue
		}
		self.db.Model(&item).Update("body", body)

		doc, err := readability.NewDocument(string(body))
		if err != nil {
			self.log.Error(err)
			continue
		}

		self.db.Model(&item).Update("readability", bluemonday.StrictPolicy().Sanitize(doc.Content()))
	}

}
