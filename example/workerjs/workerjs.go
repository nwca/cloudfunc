package workerjs

import (
	"io"
	"net/http"
	"os"
)

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		f, err := os.Open("/var/tmp/worker/worker.js")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		_, err = io.Copy(w, f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
