package main

import (
	"log"
	"net/http"
	"os"

	"sevki.org/cloud/reloader"
)

var (
	user    = os.Getenv("USER")
	repo    = os.Getenv("REPO")
	repourl = os.Getenv("REPO_URL")
	token   = os.Getenv("TOKEN")
)

func main() {

	http.Handle("/", reloader.Setup(user, repo, repourl, token, NewWiki))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
