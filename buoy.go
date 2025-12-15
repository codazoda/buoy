package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed assets/index.html
var defaultIndex []byte

//go:embed assets/com.joeldare.buoy.plist
var launchdPlistTemplate string

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

	if os.Getenv("BUOY_LAUNCHD") == "" {
		if promptAutoStart() {
			if err := installLaunchd(); err != nil {
				fmt.Printf("Failed to install launchd service: %v\n", err)
			} else {
				fmt.Println("Configured launchd to start Buoy automatically.")
			}
		}
	}

	fileServer := http.FileServer(http.Dir(wwwDir))
	http.Handle("/", fileServer)
	http.HandleFunc("/search", searchHandler)
	fmt.Printf("Serving %s on port %s\n", wwwDir, port)
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

func promptAutoStart() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Would you like Buoy to start automatically? [Y/n]: ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)
	return response == "" || strings.EqualFold(response, "y") || strings.EqualFold(response, "yes")
}

func installLaunchd() error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return err
	}

	userHome, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	plistDir := filepath.Join(userHome, "Library", "LaunchAgents")
	if err := os.MkdirAll(plistDir, 0o755); err != nil {
		return err
	}

	plistPath := filepath.Join(plistDir, "com.joeldare.buoy.plist")
	plistContent := []byte(strings.ReplaceAll(launchdPlistTemplate, "{{EXEC_PATH}}", execPath))

	if err := os.WriteFile(plistPath, plistContent, 0o644); err != nil {
		return err
	}

	// Attempt to load immediately; if it fails, leave the plist in place.
	cmd := exec.Command("launchctl", "load", plistPath)
	if err := cmd.Run(); err != nil {
		fmt.Printf("launchctl load failed (you may need to load manually): %v\n", err)
	}

	return nil
}
