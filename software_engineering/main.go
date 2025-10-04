package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/ping", PingHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprint(w, "pong\n")
}
