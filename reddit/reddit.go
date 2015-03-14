package reddit

import (
	"encoding/json"
	"fmt"
	"github.com/bnadland/newsstream/item"
	"io/ioutil"
	"net/http"
	"time"
)

func Get(items chan<- item.Item, name string) error {
	type Story struct {
		Created float64
		URL     string
		Title   string
	}

	type Subreddit struct {
		Data struct {
			Children []struct {
				Data Story
			}
		}
	}

	url := fmt.Sprintf("http://www.reddit.com/r/%s.json", name)
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var subreddit Subreddit
	err = json.Unmarshal(body, &subreddit)
	if err != nil {
		return err
	}

	for _, child := range subreddit.Data.Children {
		item := item.Item{
			Source: fmt.Sprintf("/r/%s", name),
			Title:  child.Data.Title,
			Link:   child.Data.URL,
			Date:   time.Unix(int64(child.Data.Created), 0),
		}
		items <- item
	}

	return nil
}
