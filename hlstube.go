package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/ReneKroon/ttlcache"
)

type HLSTube struct {
	m3us      *ttlcache.Cache
	transport *http.Transport
}

const (
	format    = "best[protocol^=m3u8]"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.101 Safari/537.36"
)

func NewHLSTube() *HLSTube {
	cache := ttlcache.NewCache()
	cache.SetTTL(time.Duration(5 * time.Hour))
	cache.SkipTtlExtensionOnHit(true)
	return &HLSTube{
		m3us: cache,
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

func yt2m3u(u string) (s string, err error) {
	attempts := 3
	for i := 0; i < attempts; i++ {
		if i > 1 {
			log.Printf("retrying %s\n", u)
			time.Sleep(time.Duration(i) * time.Second)
		}
		m3u, err := exec.Command("yt-dlp", "--format", format, u, "-g").Output()
		if len(m3u) > 0 && err == nil {
			fmt.Printf("%s is %s\n", u, m3u)
			trimmed := strings.TrimSpace(string(m3u))
			return trimmed, nil
		}
	}
	if err == nil {
		err = fmt.Errorf("yt-dlp failed")
	}
	return "", err
}

func (h *HLSTube) handler(w http.ResponseWriter, r *http.Request) {
	u, err := extractURL(r)
	if err != nil {
		err404(w, r)
		return
	}
	ytUrl := u.String()

	var m3u string

	value, exists := h.m3us.Get(ytUrl)

	if exists {
		m3u = value.(string)
	} else {
		log.Printf("%s wants to stream %s\n", r.RemoteAddr, ytUrl)
		m3u, err = yt2m3u(ytUrl)

		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				log.Println(string(exiterr.Stderr))
				err500(w, r, exiterr)
				return
			}
		}
		h.m3us.Set(ytUrl, m3u)
	}

	log.Printf("streaming %s to %s\n", ytUrl, r.RemoteAddr)
	http.Redirect(w, r, m3u, http.StatusFound)
}
