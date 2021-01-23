package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func logStacktrace(err error) {
	if err, ok := err.(stackTracer); ok {
		for _, f := range err.StackTrace() {
			log.Printf("%+s:%d\n", f, f)
		}
	}
}

func extractURL(r *http.Request) (*url.URL, error) {
	vars := mux.Vars(r)
	if vars["v"] != "" {
		return url.Parse(fmt.Sprintf("https://www.youtube.com/watch?v=%s", vars["v"]))
	}
	if vars["_u"] != "" {
		url, err := url.Parse(strings.Split(r.URL.Path, "/_/")[1])
		log.Printf("%#v\n", r.URL)
		url.RawQuery = r.URL.RawQuery
		if err != nil {
			return nil, err
		}
		return url, nil
	}
	return nil, errors.New("not found")
}

func errorHandler(w http.ResponseWriter, r *http.Request, err error, code int) {
	if code > 499 {
		logStacktrace(err)
	}
	http.Error(w, err.Error(), code)
}

func err404(w http.ResponseWriter, r *http.Request) {
	errorHandler(w, r, errors.New("not found"), http.StatusNotFound)
}
func err500(w http.ResponseWriter, r *http.Request, err error) {
	errorHandler(w, r, err, http.StatusInternalServerError)
}
