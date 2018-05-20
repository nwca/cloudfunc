package main

import "net/http"

func HandleHTTP(fnc http.HandlerFunc) {
	http.HandleFunc("/", fnc)
}
