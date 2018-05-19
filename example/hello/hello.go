package hello

import (
	"encoding/json"
	"net/http"
)

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"msg": "hello from Go!",
			"url": r.URL.String(),
		})
	})
}
