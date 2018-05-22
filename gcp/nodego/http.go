package main

import (
	"net/http"
	"strings"
)

func HandleHTTP(fnc http.HandlerFunc) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, executePrefix)
		fnc(w, r)
	})
}
