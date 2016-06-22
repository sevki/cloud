package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/iand/feedparser"
	"github.com/koding/cache"

	"html/template"

	"sevki.org/lib/prettyprint"
)

var (
	httpCache *cache.MemoryTTL
)

func init() {
	httpCache = cache.NewMemoryWithTTL(time.Minute)
}

var funcMap = template.FuncMap{
	"gover":      gover,
	"now":        now,
	"uptime":     uptime,
	"atom":       getAtom,
	"json":       getJson,
	"jsonxsfr":   getJsonXSFR,
	"contains":   contains,
	"isrepeated": isRepeated,
	"lessthan":   lessthan,
	"multiply":   multiply,
	"add":        add,
	"jsondate":   jsondate,
	"linkify":    linkify,
	"ppJson":     prettyprint.AsJSON,
	"regexMatch": regexMatch,
}

func regexMatch(a, b string) string {
	if re, err := regexp.Compile(a); err != nil {
		return err.Error()
	} else {
		return re.FindString(b)
	}
}
func jsondate(a string) time.Time {
	if t, err := time.Parse(time.RFC3339, a); err != nil {
		return time.Now()
	} else {
		return t
	}
}

func lessthan(a, b float64) bool    { return a < b }
func add(a, b float64) float64      { return a + b }
func multiply(a, b float64) float64 { return a * b }

func uptime() string {
	return time.Since(start).String()
}

func renderTime() string {
	return ""
}

func gover() string {
	return runtime.Version()
}

func now() time.Time {
	return time.Now()
}

func contains(a []interface{}, b interface{}) bool {

	for _, i := range a {
		if i == b {
			return true
		}
	}
	return false
}

func isRepeated(a ...string) bool {
	for _, b := range a[1:] {
		if a[0] == b {
			return true
		}
	}
	return false
}

func getAtom(url string) interface{} {
	if val, _ := httpCache.Get(url); val != nil {
		return val
	}

	resp, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
		return nil
	}

	feed, _ := feedparser.NewFeed(resp.Body)
	resp.Body.Close()
	httpCache.Set(url, feed.Items)
	return feed.Items
}

func getJsonWithXSFR(url string) interface{} {

	if val, _ := httpCache.Get(url); val != nil {
		return val
	}
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	defer resp.Body.Close()
	// The JSON response begins with an XSRF-defeating header
	// like ")]}\n". Read that and skip it.
	br := bufio.NewReader(resp.Body)
	if _, err := br.ReadSlice('\n'); err != nil {
		return err
	}

	dec := json.NewDecoder(br)
	var result interface{}

	err = dec.Decode(&result)
	if err != nil {
		log.Fatal(err)
		return nil
	} else {
		httpCache.Set(url, result)
		return result
	}
}
func getJson(url string) interface{} {

	if val, _ := httpCache.Get(url); val != nil {
		return val
	}
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal(err)
		return nil
	}
	dec := json.NewDecoder(resp.Body)
	var result interface{}

	err = dec.Decode(&result)
	if err != nil {
		log.Fatal(err)
		return nil
	} else {
		httpCache.Set(url, result)
		return result
	}
}

func getJsonXSFR(url string) interface{} {

	if val, _ := httpCache.Get(url); val != nil {
		return val
	}
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal(err)
		return nil
	}
	br := bufio.NewReader(resp.Body)
	// For security reasons or something, this URL starts with ")]}'\n" before
	// the JSON object. So ignore that.
	// Shawn Pearce says it's guaranteed to always be just one line, ending in '\n'.
	for {
		b, err := br.ReadByte()
		if err != nil {
			return nil
		}
		if b == '\n' {
			break
		}
	}

	dec := json.NewDecoder(br)
	var result interface{}

	err = dec.Decode(&result)
	if err != nil {
		log.Fatal(err)
		return nil
	} else {
		httpCache.Set(url, result)
		return result
	}
}

func linkify(s string) template.HTML {
	ss := strings.Split(s, " ")
	var sn []string
	for _, x := range ss {

		u, err := url.Parse(x)
		if err == nil && u.IsAbs() {
			sn = append(sn,
				fmt.Sprintf("<a href=\"%s\">%s</a>",
					x,
					strings.TrimLeft(x, u.Scheme+"://")))
		} else {
			sn = append(sn, x)
		}
	}

	return template.HTML(strings.Join(sn, " "))
}
