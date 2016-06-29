package main

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"rsc.io/cloud"
	"rsc.io/cloud/diskcache"
	"rsc.io/cloud/google/gcs"
)

var (
	start = time.Now()
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

	http.Handle("/", mimeTypeHandler(fs))
	http.HandleFunc("/_status", status)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
func mimeTypeHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		mimetype := "application/octet-stream"
		ext := path.Ext(r.URL.Path)
		if strings.HasSuffix(r.URL.Path, "/") {
			ext = ".html"
		}
		mt := mime.TypeByExtension(ext)
		if mt != "" {
			mimetype = mt
		}

		w.Header().Set("Content-Type", mimetype)
		h.ServeHTTP(w, r)
	})
}

func status(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK\n"))
	w.Write([]byte(fmt.Sprintf("go:\t%s\n", runtime.Version())))
	w.Write([]byte(fmt.Sprintf("uptime:\t%s\n", time.Since(start).String())))

}
