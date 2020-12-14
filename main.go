package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	hlstube := &HLSTube{}
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	hlstube.paths = make(map[string]string)
	http.HandleFunc("/", hlstube.handler)
	http.ListenAndServe(":8080", nil)
}
