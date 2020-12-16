package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	hlstube := NewHLSTube()
	r := mux.NewRouter().SkipClean(true).UseEncodedPath()
	r.HandleFunc("/_/{_u:.+}", hlstube.handler)
	r.HandleFunc("/favicon.ico", hlstube.err404)
	r.HandleFunc("/{v}", hlstube.handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("hi")
	http.ListenAndServe(fmt.Sprintf(":%s", port), r)
	log.Println("bye")
}
