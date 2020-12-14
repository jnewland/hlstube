package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"strings"
)

type HLSTube struct {
	m3us map[string]string
}

func (h *HLSTube) handler(w http.ResponseWriter, r *http.Request) {
	// TODO do some additional sanitization
	v := strings.Split(r.URL.Path, "/")[1]
	if v == "favicon.ico" {
		http.Error(w, fmt.Sprintf("%s not found", v), http.StatusNotFound)
		return
	}
	log.Println(v)
	if len(v) == 0 {
		http.Error(w, fmt.Sprintf("%s not found", v), http.StatusNotFound)
		return
	}
	if h.m3us[v] == "" {
		m3u, err := exec.Command("youtube-dl", fmt.Sprintf("https://www.youtube.com/watch?v=%s", v), "-g").Output()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.m3us[v] = strings.TrimSpace(string(m3u))
		log.Println(h.m3us[v])
	}

	origin, err := url.Parse(h.m3us[v])
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	director := func(req *http.Request) {
		req.URL.Host = origin.Host
		req.URL.Scheme = origin.Scheme
		req.URL.Path = origin.Path
		req.URL.RawPath = origin.RawPath
	}
	proxy := &httputil.ReverseProxy{Director: director}

	w.Header().Set("X-HLSTube", "is rad")
	proxy.ServeHTTP(w, r)
}
