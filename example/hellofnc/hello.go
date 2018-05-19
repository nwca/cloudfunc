package hello

import (
	"encoding/json"
	"net/http"
)

func HelloFunc(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"msg": "hello from Go func!",
		"url": r.URL.String(),
	})
}
