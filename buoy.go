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
	launchdLabel       = "com.joeldare.Buoy"
	legacyLaunchdLabel = "com.joeldare.buoy"
	defaultPort        = "42869"
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
			installPort := chooseInstallPort(port)
			if err := installLaunchd(installPort); err != nil {
				fmt.Printf("Failed to install launchd service: %v\n", err)
			} else {
				if err := startLaunchdService(); err != nil {
					fmt.Printf("Configured launchd, but failed to start Buoy via launchd: %v\n", err)
				} else {
					fmt.Printf("Buoy is now managed by launchd and running in the background.\nServing %s on port %s\n", wwwDir, installPort)
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
	domain := fmt.Sprintf("gui/%d", os.Getuid())

	cleanupLegacyLaunchd(plistDir, domain)
	// Always remove any existing instance before writing a fresh plist.
	_ = exec.Command("launchctl", "bootout", domain+"/"+launchdLabel).Run()

	envVars := buildLaunchdEnv(port)
	plistContent := strings.ReplaceAll(launchdPlistTemplate, "{{EXEC_PATH}}", execPath)
	plistContent = strings.ReplaceAll(plistContent, "{{ENV_VARS}}", envVars)
	plistContent = strings.ReplaceAll(plistContent, "{{LABEL}}", launchdLabel)

	if err := os.WriteFile(plistPath, []byte(plistContent), 0o644); err != nil {
		return err
	}

	// Bootstrap the agent into the user domain.
	if err := exec.Command("launchctl", "bootstrap", domain, plistPath).Run(); err != nil {
		return fmt.Errorf("launchctl bootstrap failed: %w", err)
	}

	// Ensure it is enabled on login.
	if err := exec.Command("launchctl", "enable", domain+"/"+launchdLabel).Run(); err != nil {
		fmt.Printf("launchctl enable failed (service may still run): %v\n", err)
	}

	// Kickstart to force-restart with the freshly written plist.
	if err := exec.Command("launchctl", "kickstart", "-k", domain+"/"+launchdLabel).Run(); err != nil {
		fmt.Printf("launchctl kickstart failed (service may still start at load): %v\n", err)
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
	// Always rebuild to ensure the persistent binary matches the current source/defaults.
	cmd := exec.Command("go", "build", "-o", target, ".")
	cmd.Dir = cwd
	cmd.Env = os.Environ()

	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("go build failed: %v: %s", err, string(out))
	}

	return target, nil
}

func cleanupLegacyLaunchd(plistDir, domain string) {
	// Try to bootout any legacy-named service and remove its plist.
	_ = exec.Command("launchctl", "bootout", domain+"/"+legacyLaunchdLabel).Run()
	legacyPlist := filepath.Join(plistDir, "com.joeldare.buoy.plist")
	_ = os.Remove(legacyPlist)
}

func chooseInstallPort(current string) string {
	envPort := os.Getenv("PORT")
	if envPort != "" && envPort != defaultPort {
		fmt.Printf("Detected PORT=%s in your environment. Use this for Buoy auto-start? [y/N]: ", envPort)
		reader := bufio.NewReader(os.Stdin)
		resp, _ := reader.ReadString('\n')
		resp = strings.TrimSpace(resp)
		if strings.EqualFold(resp, "y") || strings.EqualFold(resp, "yes") {
			return envPort
		}
		fmt.Printf("Using default port %s for Buoy auto-start.\n", defaultPort)
		return defaultPort
	}
	return current
}
