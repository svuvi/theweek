package main

import (
	"fmt"
	"log"
	"net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Hello</h1>")
}

func main() {
	http.HandleFunc("/", hello)
	log.Print("Serving localhost:8080")
	http.ListenAndServe(":8080", nil)
}
