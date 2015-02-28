package main

import (
	"encoding/json"
	"fmt"
	"github.com/caser/gophernews"
	"html/template"
	"io/ioutil"
	"net/http"
)

var allStories map[string]*Story

type Subreddit struct {
	Data struct {
		Children []struct {
			Data Story
		}
	}
}

type Story struct {
	URL   string
	Title string
}

func getHN(stories chan *Story) error {
	client := gophernews.NewClient()

	storyIds, err := client.GetTop100()
	if err != nil {
		return err
	}

	for _, storyId := range storyIds {
		go func(storyId int) {
			story, err := client.GetStory(storyId)
			if err != nil {
				return
			}
			stories <- &Story{
				Title: story.Title,
				URL:   story.URL,
			}
		}(storyId)
	}

	return nil
}

func getSubreddit(stories chan *Story, name string) error {
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
		stories <- &Story{
			Title: child.Data.Title,
			URL:   child.Data.URL,
		}
	}

	return nil
}

func indexPage(w http.ResponseWriter, req *http.Request) {
	templateString := `
	{{range .}}
		<a href="{{.URL}}">{{.Title}}</a><br>
	{{end}}
`
	t, _ := template.New("index").Parse(templateString)
	t.Execute(w, allStories)
}

func main() {

	stories := make(chan *Story)

	go getHN(stories)

	for _, name := range []string{
		"golang",
		"programming",
		"webdev",
		"python",
	} {
		go getSubreddit(stories, name)
	}

	go func() {
		allStories = make(map[string]*Story)
		for {
			story := <-stories
			allStories[story.URL] = story
		}
	}()

	http.Handle("/", http.HandlerFunc(indexPage))
	http.ListenAndServe(":8080", nil)
}
