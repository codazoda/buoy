package main

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

//go:embed assets/index.html
var defaultIndex []byte

const (
	defaultPort = "42869"
)

func main() {
	port := getServerPort()
	wwwDir, err := getWWWDir()
	if err != nil {
		fmt.Printf("Could not determine www directory: %v\n", err)
		os.Exit(1)
	}

	if err := ensureIndexFile(wwwDir); err != nil {
		fmt.Printf("Could not initialize index.html: %v\n", err)
		os.Exit(1)
	}

	fileServer := http.FileServer(http.Dir(wwwDir))
	http.Handle("/", fileServer)
	http.HandleFunc("/search", searchHandler)
	fmt.Printf("Serving %s on port %s\n", wwwDir, port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println(err.Error())
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	io.WriteString(w, "Searched for "+query+"\n")
}

func getServerPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	return port
}

func getWWWDir() (string, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataHome = filepath.Join(userHome, ".local", "share")
	}

	wwwDir := filepath.Join(dataHome, "buoy", "www")
	if err := os.MkdirAll(wwwDir, 0o755); err != nil {
		return "", err
	}

	return wwwDir, nil
}

func ensureIndexFile(wwwDir string) error {
	indexPath := filepath.Join(wwwDir, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	return os.WriteFile(indexPath, defaultIndex, 0o644)
}
