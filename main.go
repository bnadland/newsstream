package main

import (
	"encoding/json"
	"fmt"
	"github.com/caser/gophernews"
	"gopkg.in/antage/eventsource.v1"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
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
	var knownStoryIds []int

	client := gophernews.NewClient()

	for {
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
				knownStoryIds = append(knownStoryIds, storyId)
				stories <- &Story{
					Title: story.Title,
					URL:   story.URL,
				}
			}(storyId)
		}

		time.Sleep(time.Hour*3 + time.Minute*time.Duration(rand.Intn(60)))
	}

	return nil
}

func getSubreddit(stories chan *Story, name string) error {
	for {
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

		time.Sleep(time.Hour*3 + time.Minute*time.Duration(rand.Intn(60)))
	}

	return nil
}

func indexPage(w http.ResponseWriter, req *http.Request) {
	templateString := `<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title>Newsstream</title>
	<link rel="stylesheet" href="//cdnjs.cloudflare.com/ajax/libs/animate.css/3.2.1/animate.min.css">
	<style>
	li { list-style-type: none; }
	li:before { content: "> "; }
	</style>
	<script src="//cdnjs.cloudflare.com/ajax/libs/jquery/2.1.3/jquery.min.js"></script>
	<script>
 $(function () {
        var evsrc = new EventSource("/newlinks")
        evsrc.onmessage = function (ev) {
		  data = $.parseJSON(ev.data)
		  console.log(data)
          $("#links").prepend("<li><a class=\"link animated fadeIn\" href=" + data.URL + ">" + data.Title + "</a></li>")
        }
        evsrc.onerror = function (ev) {
          console.log("readyState = " + ev.currentTarget.readyState)
        }
      })
	</script>
</head>
<body>
	<ul id="links">
	{{range .}}
		<li><a class="link" href="{{.URL}}">{{.Title}}</a></li>
	{{end}}
	</ul>
</body>
</html>
`
	t, _ := template.New("index").Parse(templateString)
	t.Execute(w, allStories)
}

func main() {

	stories := make(chan *Story)

	es := eventsource.New(nil, nil)
	defer es.Close()

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
			storyJson, err := json.Marshal(story)
			if err != nil {
				fmt.Println(err)
			} else {
				es.SendEventMessage(string(storyJson), "", "")
				time.Sleep(time.Second)
			}
		}
	}()

	http.Handle("/newlinks", es)
	http.Handle("/", http.HandlerFunc(indexPage))
	http.ListenAndServe(":8080", nil)
}
