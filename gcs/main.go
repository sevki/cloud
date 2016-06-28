package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"rsc.io/cloud"
	"rsc.io/cloud/diskcache"
	"rsc.io/cloud/google/gcs"
)

func main() {
	if !strings.HasPrefix(os.Getenv("BUCKET"), "gs://") {
		log.Fatal("-webroot argument must be a gs:// URL")
	}
	loader, err := gcs.NewLoader("/")
	if err != nil {
		log.Fatal(err)
	}

	cache, err := diskcache.New("./gcscache", loader)
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(cloud.Dir(cache, strings.TrimPrefix(os.Getenv("BUCKET"), "gs://")))
	http.Handle("/", fs)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
