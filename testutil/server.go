package testutil

import (
	"net/http"
	"strconv"
)

func TestServer(port int) {
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte{1, 2})
	})
	http.ListenAndServe(":"+strconv.Itoa(port), handler)
}
