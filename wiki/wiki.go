package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"html/template"
)

var (
	start = time.Now()
)

type WikiServer struct {
	dir string
}

func (ws *WikiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	page := make(map[string]interface{})

	page["requestStart"] = time.Now()
	page["start"] = start
	page["url"] = r.URL.String()
	page["host"] = r.Host

	root := filepath.Join(ws.dir, "root.tmpl")
	parse := func(name string) (*template.Template, error) {
		t := template.New("root").Funcs(funcMap)
		p := filepath.Join(ws.dir, name)

		return t.ParseFiles(root, p)
	}
	pageName := fmt.Sprintf("%s.tmpl", r.URL.Path)
	if strings.TrimRight(r.URL.Path, "/") != r.URL.Path {
		pageName = filepath.Join(r.URL.Path, "home.tmpl")
	}

	if tmpl, err := parse(pageName); err == nil {
		tmpl.ExecuteTemplate(w, "root", page)
	} else if os.IsNotExist(err) {
		log.Println(err)
		http.ServeFile(w, r, filepath.Join(ws.dir, r.URL.Path))
	} else {
		http.Error(w, err.Error(), 500)
	}
}

func NewWiki(path string) http.Handler {

	wiki := WikiServer{
		dir: path,
	}

	return &wiki
}