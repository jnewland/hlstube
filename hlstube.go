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

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type HLSTube struct {
	m3us      map[string]string
	transport *http.Transport
}
type stackTracer interface {
	StackTrace() errors.StackTrace
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

func logStacktrace(err error) {
	if err, ok := err.(stackTracer); ok {
		for _, f := range err.StackTrace() {
			log.Printf("%+s:%d\n", f, f)
		}
	}
}

func extractUrl(r *http.Request) (*url.URL, error) {
	vars := mux.Vars(r)
	if vars["v"] != "" {
		return url.Parse(fmt.Sprintf("https://www.youtube.com/watch?v=%s", vars["v"]))
	}
	if vars["_u"] != "" {
		url, err := url.Parse(strings.Split(r.URL.Path, "/_/")[1])
		url.RawQuery = r.URL.RawQuery
		if err != nil {
			return nil, err
		}
		return url, nil
	}
	return nil, errors.New("not found")
}

func (h *HLSTube) errorHandler(w http.ResponseWriter, r *http.Request, err error, code int) {
	if code > 499 {
		logStacktrace(err)
	}
	http.Error(w, err.Error(), code)
}

func (h *HLSTube) err404(w http.ResponseWriter, r *http.Request) {
	h.errorHandler(w, r, errors.New("not found"), http.StatusNotFound)
}
func (h *HLSTube) err500(w http.ResponseWriter, r *http.Request, err error) {
	h.errorHandler(w, r, err, http.StatusInternalServerError)
}

func (h *HLSTube) handler(w http.ResponseWriter, r *http.Request) {
	u, err := extractUrl(r)
	if err != nil {
		h.err404(w, r)
		return
	}
	if h.m3us[u.String()] == "" {
		log.Printf("setting up a stream for %s\n", u.String())
		m3u, err := exec.Command("youtube-dl", u.String(), "-g").Output()
		if err != nil {
			h.err500(w, r, err)
			return
		}
		h.m3us[u.String()] = strings.TrimSpace(string(m3u))
	}

	origin, err := url.Parse(h.m3us[u.String()])
	if err != nil {
		h.err500(w, r, err)
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
	proxy := &httputil.ReverseProxy{Director: director, ModifyResponse: modifyResponse, Transport: h.transport, ErrorHandler: h.err500}

	w.Header().Set("X-HLSTube", "is rad")
	proxy.ServeHTTP(w, r)
}
