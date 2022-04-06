package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	hlsTube := NewHLSTube()
	m3uTube := NewM3UTube()
	ffmpegHandler := NewFFmpegHandler()
	r := mux.NewRouter().SkipClean(true).UseEncodedPath()
	r.HandleFunc("/_p/{_p:.+}", m3uTube.handler)
	r.HandleFunc("/_/{_u:.+}", hlsTube.handler)
	r.HandleFunc("/_f/{url:.+}", ffmpegHandler.PlaylistHandler)
	r.HandleFunc("/_s/{stream}/{segment}", ffmpegHandler.SegmentHandler)
	go ffmpegHandler.ExpireStaleStreams()
	r.HandleFunc("/favicon.ico", err404)
	r.HandleFunc("/{v}", hlsTube.handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("hi")
	http.ListenAndServe(fmt.Sprintf(":%s", port), handlers.ProxyHeaders(r))
	log.Println("bye")
}
