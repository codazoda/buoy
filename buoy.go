package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	port := getServerPort()
	fileServer := http.FileServer(http.Dir("./www"))
	http.Handle("/", fileServer)
	http.HandleFunc("/search", searchHandler)
	fmt.Printf("Server started on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf(err.Error())
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	io.WriteString(w, "Searched for "+query+"\n")
}

func getServerPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}
