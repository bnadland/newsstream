package item

import (
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
