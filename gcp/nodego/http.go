package main

import (
	"log"
	"net/http"
	"os"
	"strings"
)

var fixRedirects = os.Getenv("CLOUDFUNC_FIX_REDIRECTS") == "true"

func HandleHTTP(fnc http.HandlerFunc) {
	var h http.Handler = fnc
	h = http.StripPrefix(executePrefix, h)
	if fixRedirects {
		h = redirectHandler{h}
	}
	http.Handle("/", WithLogger(h))
}

type redirectHandler struct {
	h http.Handler
}

func (h redirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.h.ServeHTTP(redirectWriter{w}, r)
}

type redirectWriter struct {
	w http.ResponseWriter
}

func (w redirectWriter) Header() http.Header {
	return w.w.Header()
}

func (w redirectWriter) Write(p []byte) (int, error) {
	return w.w.Write(p)
}

func (w redirectWriter) WriteHeader(statusCode int) {
	if statusCode == http.StatusMovedPermanently {
		h := w.w.Header()
		const hdr = "Location"
		loc := h.Get(hdr)
		if pref := "/" + functionName; !strings.HasPrefix(loc, pref) {
			loc = pref + loc
			h.Set(hdr, loc)
		}
		log.Println("redirect:", loc)
	}
	w.w.WriteHeader(statusCode)
}
