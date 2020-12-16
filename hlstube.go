package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

type HLSTube struct {
	m3us      map[string]string
	transport *http.Transport
}

func NewHLSTube() *HLSTube {
	return &HLSTube{
		m3us: make(map[string]string),
		transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: false,
			}).DialContext,
			ForceAttemptHTTP2:     false,
			MaxIdleConns:          10,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   30 * time.Second,
			ExpectContinueTimeout: 10 * time.Second,
		},
	}
}

func (h *HLSTube) handler(w http.ResponseWriter, r *http.Request) {
	// TODO do some additional sanitization
	v := strings.Split(r.URL.Path, "/")[1]
	if v == "favicon.ico" {
		http.Error(w, fmt.Sprintf("%s not found", v), http.StatusNotFound)
		return
	}
	if len(v) == 0 {
		http.Error(w, fmt.Sprintf("%s not found", v), http.StatusNotFound)
		return
	}
	if h.m3us[v] == "" {
		log.Printf("setting up a stream for %s\n", v)
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
		req.Host = origin.Host
		req.URL.Host = origin.Host
		req.URL.Scheme = origin.Scheme
		req.URL.Path = origin.Path
		req.URL.RawPath = origin.RawPath
	}
	modifyResponse := func(resp *http.Response) error {
		log.Printf("%s %d\n", v, resp.StatusCode)
		if resp.StatusCode != http.StatusOK {
			h.m3us[v] = ""
			log.Printf("forgot about %s\n", v)
			resp.Header.Set("X-HLSTube-reset", "m3u forgotten, try again")
		}
		return nil
	}
	proxy := &httputil.ReverseProxy{Director: director, ModifyResponse: modifyResponse, Transport: h.transport}

	w.Header().Set("X-HLSTube", "is rad")
	proxy.ServeHTTP(w, r)
}
