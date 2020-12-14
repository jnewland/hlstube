package main

import (
	"net/http"
)

func main() {
	hlstube := &HLSTube{}
	hlstube.m3us = make(map[string]string)
	http.HandleFunc("/", hlstube.handler)
	http.ListenAndServe(":8080", nil)
}
