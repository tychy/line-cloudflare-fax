package main

import (
	"io"
	"net/http"

	"github.com/syumai/workers"
)

func main() {
	http.HandleFunc("/hello", func(w http.ResponseWriter, req *http.Request) {
		msg := "Hello!"
		w.Write([]byte(msg))
	})
	http.HandleFunc("/echo", func(w http.ResponseWriter, req *http.Request) {
		io.Copy(w, req.Body)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		msg := "Top Page"
		w.Write([]byte(msg))
	})
	workers.Serve(nil) // use http.DefaultServeMux
}
