package main

import (
	"fmt"
	"net/http"
)

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(message))
}

func upload(w http.ResponseWriter, r *http.Request) {
	// parse and validate file and post parameters
	file, _, err := r.FormFile("file")
	if err != nil {
		renderError(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}

	path := r.FormValue("path")
	if err != nil {
		renderError(w, "INVALID_PATH", http.StatusBadRequest)
		return
	}

	fmt.Println(path)
	fmt.Println(file)
}

func main() {
	http.HandleFunc("/upload", upload)
	http.ListenAndServe(":8080", nil)
}
