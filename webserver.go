package main

import (
	"github.com/dustin/go-humanize"
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

	tmplFuncs := make(map[string]interface{})
	tmplFuncs["formatTime"] = humanize.Time
	tmpl, err := template.New("index").Funcs(tmplFuncs).Parse(`<!doctype html><html>
<head>
	<meta charset="utf-8" />
	<meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0">
	<title>newsstream</title>
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/semantic-ui/1.11.8/semantic.min.css"/>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/2.1.3/jquery.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/semantic-ui/1.11.8/semantic.min.js"></script>
</head>
</head>
<body>
<div class="ui main text container">
	<div class="ui segment">
		<div class="ui relaxed divided list">
		{{ range . }}
			<div class="item">
				<div class="content">
					<a class="header" href="{{ .Url }}">{{ .Title }}</a>
					<div class="description">{{ formatTime .CreatedAt }} - {{ .Source }}</div>
				</div>
			</div>
		{{ else }}
			<div class="item">
				<div class="content">
					<a class="header">No items.</a>
					<div class="description"></div>
				</div>
			</div>
		{{ end }}
		</div>
	</div>
</div>
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
