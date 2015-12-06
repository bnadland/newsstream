package main

import (
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"html/template"
	"net/http"
)

func (self *Newsstream) httpIndexPage(c web.C, w http.ResponseWriter, r *http.Request) {

	var items []Item
	err := self.db.Order("created_at desc").Limit(20).Find(&items).Error
	if err != nil {
		self.log.Error(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("index").Parse(`<!doctype html><html>
<head>
	<title>newsstream</title>
</head>
<body>
<ul>
{{ range . }}
	<li><a href="{{ .Url }}">{{ .Title }}</a></li>
{{ else }}
	<li>No items.</li>
{{ end }}
</ul>
</body>
</html>
`)
	if err != nil {
		self.log.Error(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, items)
	if err != nil {
		self.log.Error(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func (self *Newsstream) webServer() {
	graceful.PostHook(self.Close)
	graceful.HandleSignals()

	routes := web.New()

	routes.Get("/", self.httpIndexPage)

	self.log.Info("Listening on ", self.config.Listen)
	graceful.ListenAndServe(self.config.Listen, routes)
}
