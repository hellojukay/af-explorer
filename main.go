package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"
)

var (
	port   = 8080
	prefix = "http://exmaple.com/artifactory/repo"
	suffix = map[string]string{
		".json": "application/json",
		".html": "text/html; charset=utf-8",
		".txt":  "text/plain; charset=utf-8",
		".xml":  "application/xml; charset=utf-8",
		".pdf":  "application/pdf; charset=utf-8",
		".gif":  "application/gif; charset=utf-8",
		".jpe":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".js":   "text/javascript; charset=utf-8",
		".mp3":  "audio/mp3",
		".mp4":  "video/mpeg4",
		".css":  "text/css",
	}
)

func main() {
	flag.IntVar(&port, "port", port, "http port")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url, _ := url.Parse(prefix)
		proxy := httputil.ReverseProxy{
			Director: func(r *http.Request) {
				r.URL.Scheme = url.Scheme
				r.URL.Host = url.Host
				r.URL.Path = url.Path + r.URL.Path
				r.Host = url.Host
				if strings.HasSuffix(r.URL.Path, "/") && hasIndex(r.URL.String()) {
					r.URL.Path += "index.html"
				}
			},
			ModifyResponse: ModifyResponse,
		}
		if strings.Contains(r.URL.Path, ";base64") {
			return
		}
		proxy.ServeHTTP(w, r)

	})
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

// ModifyResponse ModifyResponse
func ModifyResponse(w *http.Response) error {
	var extension = filepath.Ext(w.Request.RequestURI)
	w.Header.Set("Access-Control-Allow-Origin", "*")
	if contenType, ok := suffix[extension]; ok {
		w.Header.Set("content-type", contenType)
	}
	w.Header.Del(`Strict-Transport-Security`)
	w.Header.Del(`Content-Security-Policy`)
	for k := range w.Header {
		if strings.HasPrefix(k, "X-") {
			w.Header.Del(k)
		}
	}
	log.Printf("%s %d\n", w.Request.URL.String(), w.StatusCode)

	return nil
}

func hasIndex(path string) bool {
	if strings.HasSuffix(path, "/") {
		path += "index.html"
	} else {
		path += "/index.html"
	}
	res, err := http.Get(path)
	if err != nil {
		return false
	}
	log.Printf("%s %d\n", path, res.StatusCode)

	return res.StatusCode == 200
}
