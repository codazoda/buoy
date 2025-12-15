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

const (
	launchdLabel = "com.joeldare.Buoy"
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

	if os.Getenv("BUOY_LAUNCHD") == "" {
		if promptAutoStart() {
			if err := installLaunchd(port); err != nil {
				fmt.Printf("Failed to install launchd service: %v\n", err)
			} else {
				if err := startLaunchdService(); err != nil {
					fmt.Printf("Configured launchd, but failed to start Buoy via launchd: %v\n", err)
				} else {
					fmt.Printf("Buoy is now managed by launchd and running in the background.\nServing %s on port %s\n", wwwDir, port)
					return
				}
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

func installLaunchd(port string) error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return err
	}
	execPath, err = ensurePersistentBinary(execPath)
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

	envVars := buildLaunchdEnv(port)
	plistContent := strings.ReplaceAll(launchdPlistTemplate, "{{EXEC_PATH}}", execPath)
	plistContent = strings.ReplaceAll(plistContent, "{{ENV_VARS}}", envVars)
	plistContent = strings.ReplaceAll(plistContent, "{{LABEL}}", launchdLabel)

	if err := os.WriteFile(plistPath, []byte(plistContent), 0o644); err != nil {
		return err
	}

	domain := fmt.Sprintf("gui/%d", os.Getuid())

	// Remove any existing instance, ignore errors (likely not present).
	_ = exec.Command("launchctl", "bootout", domain+"/"+launchdLabel).Run()

	// Bootstrap the agent into the user domain.
	if err := exec.Command("launchctl", "bootstrap", domain, plistPath).Run(); err != nil {
		return fmt.Errorf("launchctl bootstrap failed: %w", err)
	}

	// Ensure it is enabled on login.
	if err := exec.Command("launchctl", "enable", domain+"/"+launchdLabel).Run(); err != nil {
		fmt.Printf("launchctl enable failed (service may still run): %v\n", err)
	}

	return nil
}

func startLaunchdService() error {
	domain := fmt.Sprintf("gui/%d", os.Getuid())
	cmd := exec.Command("launchctl", "kickstart", "-k", domain+"/"+launchdLabel)
	return cmd.Run()
}

func buildLaunchdEnv(port string) string {
	var b strings.Builder
	b.WriteString("\t\t<key>BUOY_LAUNCHD</key>\n\t\t<string>1</string>\n")
	if port != "" {
		b.WriteString(fmt.Sprintf("\t\t<key>PORT</key>\n\t\t<string>%s</string>\n", port))
	}
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		b.WriteString(fmt.Sprintf("\t\t<key>XDG_DATA_HOME</key>\n\t\t<string>%s</string>\n", xdg))
	}
	return b.String()
}

func ensurePersistentBinary(execPath string) (string, error) {
	if strings.Contains(execPath, string(filepath.Separator)+"go-build"+string(filepath.Separator)) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}

		userHome, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		targetDir := filepath.Join(userHome, ".local", "bin")
		if err := os.MkdirAll(targetDir, 0o755); err != nil {
			return "", err
		}

		target := filepath.Join(targetDir, "Buoy")
		cmd := exec.Command("go", "build", "-o", target, ".")
		cmd.Dir = cwd
		cmd.Env = os.Environ()

		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("go build failed: %v: %s", err, string(out))
		}

		return target, nil
	}

	return execPath, nil
}
