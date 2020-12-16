package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"

	"github.com/gorilla/mux"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type M3UTube struct {
	streams []*stream
}

type stream struct {
	name string
	url  string
}

func NewM3UTube() *M3UTube {
	return &M3UTube{
		streams: make([]*stream, 0),
	}
}

func (h *M3UTube) handler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	playlistJson, err := exec.Command("youtube-dl", "-j", "--flat-playlist", "-i", vars["_p"]).Output()
	if err != nil {
		err500(w, r, err)
		return
	}
	w.Header().Set("X-M3UTube", "is rad")
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "#EXTM3U\n\n")

	for _, json := range strings.Split(string(playlistJson), "\n") {
		title := jsoniter.Get([]byte(json), "title").ToString()
		id := jsoniter.Get([]byte(json), "id").ToString()
		if id == "" || title == "" {
			continue
		}

		extinf := fmt.Sprintf(`#EXTINF:-1 tvg-logo="https://i.ytimg.com/vi/%s/maxresdefault.jpg" tvc-guide-art="https://i.ytimg.com/vi/%s/maxresdefault.jpg" tvg-name="%s" tvc-guide-title="%s",%s`, id, id, title, title, title)
		io.WriteString(w, extinf)
		io.WriteString(w, "\n")

		rewrittenURL, err := url.Parse(r.URL.String())
		if err != nil {
			logStacktrace(err)
			continue
		}
		rewrittenURL.Host = r.Host
		// actually set by ProxyHeaders
		rewrittenURL.Scheme = r.URL.Scheme
		if rewrittenURL.Scheme == "" {
			// default to http
			rewrittenURL.Scheme = "http"
		}
		rewrittenURL.Path = fmt.Sprintf("/%s", id)

		io.WriteString(w, rewrittenURL.String())
		io.WriteString(w, "\n")
	}

}
