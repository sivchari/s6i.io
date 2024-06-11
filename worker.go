//go:build js && wasm

package main

import (
	"embed"
	"fmt"
	"html/template"
	"iter"
	"net/http"
	"strings"

	"github.com/syumai/workers"
)

//go:embed pages
var fs embed.FS

type Worker struct{}

func (w *Worker) Work() {
	for k, v := range handlers() {
		http.HandleFunc(k, v)
	}
	workers.Serve(nil)
}

func handlers() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"GET /healthz": healthz(),
		"GET /":        index(),
		"GET /{page}":  page(),
	}
}

func healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "ok")
	}
}

func index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := fs.ReadFile("pages/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.New("index").Parse(string(p)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		entries, err := fs.ReadDir("pages")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var ps []Page
		for name := range pages(entries) {
			ps = append(ps, Page{
				URL:   strings.TrimSuffix(name, ".html"),
				Title: strings.TrimSuffix(strings.ReplaceAll(name, "_", " "), ".html"),
			})
		}

		w.Header().Set("Content-Type", "text/html")
		tmpl.Execute(w, ps)
	}
}

func page() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := r.PathValue("page")
		p, err := fs.ReadFile(fmt.Sprintf("pages/%s.html", title))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write(p)
	}
}

type Page struct {
	URL   string
	Title string
}

func pages[T interface{ Name() string }](entry []T) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, page := range entry {
			name := page.Name()
			if !strings.HasSuffix(name, ".html") || name == "index.html" {
				continue
			}
			if !yield(name) {
				return
			}
		}
	}
}
