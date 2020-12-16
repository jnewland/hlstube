package main

import (
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
	u, err := extractUrl(r)
	if err != nil {
		err404(w, r)
		return
	}
	if h.m3us[u.String()] == "" {
		log.Printf("setting up a stream for %s\n", u.String())
		m3u, err := exec.Command("youtube-dl", u.String(), "-g").Output()
		if err != nil {
			err500(w, r, err)
			return
		}
		h.m3us[u.String()] = strings.TrimSpace(string(m3u))
	}

	origin, err := url.Parse(h.m3us[u.String()])
	if err != nil {
		err500(w, r, err)
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
		log.Printf("%s %d\n", u, resp.StatusCode)
		if resp.StatusCode != http.StatusOK {
			h.m3us[u.String()] = ""
			log.Printf("forgot about %s\n", u.String())
			resp.Header.Set("X-HLSTube-reset", "m3u forgotten, try again")
		}
		return nil
	}
	proxy := &httputil.ReverseProxy{Director: director, ModifyResponse: modifyResponse, Transport: h.transport, ErrorHandler: err500}

	w.Header().Set("X-HLSTube", "is rad")
	proxy.ServeHTTP(w, r)
}
