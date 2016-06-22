package reloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	client *github.Client
	start  = time.Now()
)

type reloadFunc func(string) http.Handler 

func (srv *Server) status(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK\n"))
	w.Write([]byte(fmt.Sprintf("sha:\t%s\n", srv.sha)))
	w.Write([]byte(fmt.Sprintf("side:\t%s\n", srv.side)))
	w.Write([]byte(fmt.Sprintf("go:\t%s\n", runtime.Version())))
	w.Write([]byte(fmt.Sprintf("uptime:\t%s\n", time.Since(start).String())))

}

type Server struct {
	serving http.Handler
	err     error

	rf reloadFunc

	mu              sync.Mutex
	sha             string
	side            string
	user, repo, url string
}

func Setup(user, repo, url, ghtoken string, rf reloadFunc) *Server {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghtoken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client = github.NewClient(tc)

	srv := Server{
		user: user,
		repo: repo,
		url:  url,
		rf:   rf,
		side: "joker",
		sha:  "deadbeef",
	}
	http.HandleFunc("/_status", srv.status)

	go srv.reload()
	return &srv
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if srv.serving == nil {
		http.Error(w, "server not ready", 500)
		return
	}
	srv.serving.ServeHTTP(w, r)
}
func (srv *Server) serveHealthCheck(w http.ResponseWriter, r *http.Request) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.serving == nil {
		http.Error(w, "not ready", 500)
		return
	}
	io.WriteString(w, "ok")
}
func (srv *Server) reload() {

	for {
		srv.mu.Lock()

		// list all repositories for the authenticated user
		commits, _, err := client.Repositories.ListCommits(srv.user, srv.repo, nil)
		if err != nil {
			srv.err = fmt.Errorf("error getting commit list from github: %s", err.Error())
			log.Fatal("fetchClient :", err)
		}
		if srv.sha != *commits[0].SHA {
			newSide := "joker"
			if srv.side == "joker" {
				newSide = "batman"
			}

			newSidePath := filepath.Join("/tmp", srv.repo, newSide)

			err := checkout(srv.url, *commits[0].SHA, newSidePath)
			if err != nil {
				log.Fatal("checkout:", err)
			}

			newHandler := srv.rf(newSidePath)
			if err != nil {
				log.Fatal("rf:", err)
			}
			srv.serving = newHandler

			srv.sha = *commits[0].SHA
			srv.side = newSide

		}
		srv.mu.Unlock()
		time.Sleep(30 * time.Second)

	}
}

func checkout(repo, hash, path string) error {
	// Clone git repo if it doesn't exist.
	if _, err := os.Stat(filepath.Join(path, ".git")); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := runErr(exec.Command("git", "clone", repo, path)); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// Pull down changes and update to hash.
	cmd := exec.Command("git", "fetch")
	cmd.Dir = path
	if err := runErr(cmd); err != nil {
		return err
	}
	cmd = exec.Command("git", "reset", "--hard", hash)
	cmd.Dir = path
	if err := runErr(cmd); err != nil {
		return err
	}
	cmd = exec.Command("git", "clean", "-d", "-f", "-x")
	cmd.Dir = path
	return runErr(cmd)
}

func runErr(cmd *exec.Cmd) error {
	out, err := cmd.CombinedOutput()
	if err != nil {
		if len(out) == 0 {
			return err
		}
		return fmt.Errorf("%s\n%v", out, err)
	}
	return nil
}
