package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/ping", PingHandler)
	http.ListenAndServe(":8080", nil)
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "pong\n")
}
