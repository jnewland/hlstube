package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	hlstube := &HLSTube{}
	hlstube.m3us = make(map[string]string)
	http.HandleFunc("/", hlstube.handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
