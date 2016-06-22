package main

import (
	"flag"
	"log"
	"net/http"

	"sevki.org/cloud/reloader"
)

var (
	user  = flag.String("user", "sevki", "user")
	repo  = flag.String("repo", "cloud", "repo")
	repourl   = flag.String("url", "https://github.com/sevki/cloud.git", "repo url")
	token = flag.String("token", "", "github token")
)

func main() {
	flag.Parse()

	http.Handle("/", reloader.Setup(*user, *repo, *repourl, *token, NewWiki))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
