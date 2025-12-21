package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

//go:embed assets/index.html
var defaultIndex []byte

const (
	defaultPort         = "42869"
	defaultIndexMarker  = "<!-- Buoy default index -->"
)

var wwwRoot string

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

	wwwRoot = wwwDir
	fileServer := http.FileServer(http.Dir(wwwDir))
	http.Handle("/", fileServer)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/folders", foldersHandler)
	fmt.Printf("Serving %s on port %s\n", wwwDir, port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println(err.Error())
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	io.WriteString(w, "Searched for "+query+"\n")
}

func foldersHandler(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(wwwRoot)
	if err != nil {
		http.Error(w, "Could not read www directory", http.StatusInternalServerError)
		return
	}

	folders := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			folders = append(folders, entry.Name())
			continue
		}
		if entry.Type()&os.ModeSymlink == 0 {
			continue
		}
		info, err := os.Stat(filepath.Join(wwwRoot, entry.Name()))
		if err != nil {
			continue
		}
		if info.IsDir() {
			folders = append(folders, entry.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(folders); err != nil {
		http.Error(w, "Could not encode folder list", http.StatusInternalServerError)
	}
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
	if data, err := os.ReadFile(indexPath); err == nil {
		if !bytes.Contains(data, []byte(defaultIndexMarker)) {
			return nil
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	return os.WriteFile(indexPath, defaultIndex, 0o644)
}
