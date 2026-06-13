package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"moon/internal/config"
	"moon/internal/daemon"
)

var version = "dev"

func homeDir() string {
	return os.Getenv("HOME")
}

func xdgConfigDir() string {
	if d, err := os.UserConfigDir(); err == nil {
		return d
	}
	return filepath.Join(homeDir(), ".config")
}

func xdgDataDir() string {
	return filepath.Join(homeDir(), ".local", "share")
}

func xdgBinDir() string {
	return filepath.Join(homeDir(), ".local", "bin")
}

func cfgPath() string {
	if v := os.Getenv("MOON_CONFIG"); v != "" {
		return v
	}
	return filepath.Join(xdgConfigDir(), "moon", "config.yaml")
}

func dbDefaultPath() string {
	return filepath.Join(xdgDataDir(), "moon", "moon.db")
}

func binPath() string {
	return filepath.Join(xdgBinDir(), "moon")
}

func svcPath() string {
	return filepath.Join(xdgConfigDir(), "systemd", "user", "moon.service")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "daemon":
		cmdDaemon()
	case "start":
		cmdStart()
	case "stop":
		cmdStop()
	case "status":
		cmdStatus()
	case "install":
		cmdInstall()
	case "uninstall":
		cmdUninstall()
	case "update":
		cmdUpdate()
	case "version":
		cmdVersion()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`moon -- system monitoring daemon

Commands:
  daemon    run daemon in foreground (used by systemd)
  start     start via systemd --user
  stop      stop via systemd --user
  status    show daemon status
  install   install binary, config, and systemd user service
  uninstall remove all files
  update    update to latest version
  version   print version`)
}

func cmdDaemon() {
	if err := daemon.Run(cfgPath()); err != nil {
		log.Fatalf("daemon error: %v", err)
	}
}

func cmdStart() {
	if err := exec.Command("systemctl", "--user", "start", "moon").Run(); err != nil {
		log.Fatalf("start failed: %v", err)
	}
	fmt.Println("started")
}

func cmdStop() {
	if err := exec.Command("systemctl", "--user", "stop", "moon").Run(); err != nil {
		log.Fatalf("stop failed: %v", err)
	}
	fmt.Println("stopped")
}

func cmdStatus() {
	fmt.Printf("version: %s\n", version)

	out, _ := exec.Command("systemctl", "--user", "is-active", "moon").Output()
	switch strings.TrimSpace(string(out)) {
	case "active":
		fmt.Println("status: running")
	case "inactive":
		fmt.Println("status: stopped")
	case "failed":
		fmt.Println("status: failed")
	default:
		fmt.Println("status: not installed")
	}

	if _, err := os.Stat(svcPath()); err == nil {
		fmt.Println("autostart: enabled")
	} else {
		fmt.Println("autostart: disabled")
	}
}

func cmdInstall() {
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("executable path: %v", err)
	}

	// check if already installed
	if _, err := os.Stat(binPath()); err == nil {
		fmt.Println("moon already installed")
		fmt.Print("reinstall? [y/N] ")
		var reply string
		fmt.Scanln(&reply)
		if reply != "y" && reply != "Y" {
			fmt.Println("cancelled")
			os.Exit(0)
		}
	}

	// --- binary ---
	if err := os.MkdirAll(xdgBinDir(), 0755); err != nil {
		log.Fatalf("create bin dir: %v", err)
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		log.Fatalf("read binary: %v", err)
	}
	if err := os.WriteFile(binPath(), data, 0755); err != nil {
		log.Fatalf("write binary: %v", err)
	}
	fmt.Println("binary:", binPath())

	// --- config ---
	cfgDir := filepath.Join(xdgConfigDir(), "moon")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		log.Fatalf("create config dir: %v", err)
	}

	cfgDst := filepath.Join(cfgDir, "config.yaml")
	if _, err := os.Stat(cfgDst); os.IsNotExist(err) {
		// try config.example.yaml next to the running binary
		srcCfg := filepath.Join(filepath.Dir(exe), "config.example.yaml")
		cfgContent, readErr := os.ReadFile(srcCfg)
		if readErr != nil {
			cfgContent = []byte(fmt.Sprintf("storage:\n  db_path: %q\n", dbDefaultPath()))
		} else {
			// rewrite db_path from /root/... to user path
			cfgContent = []byte(strings.ReplaceAll(
				string(cfgContent),
				`db_path: "/root/.moon/moon.db"`,
				fmt.Sprintf(`db_path: %q`, dbDefaultPath()),
			))
		}
		if err := os.WriteFile(cfgDst, cfgContent, 0644); err != nil {
			log.Fatalf("write config: %v", err)
		}
		fmt.Println("config:", cfgDst)
	} else {
		fmt.Println("config exists, skip")
	}

	// --- data dir ---
	if err := os.MkdirAll(xdgDataDir()+"/moon", 0755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}

	// --- systemd user unit ---
	svcDir := filepath.Dir(svcPath())
	if err := os.MkdirAll(svcDir, 0755); err != nil {
		log.Fatalf("create systemd user dir: %v", err)
	}

	content := fmt.Sprintf(`[Unit]
Description=Moon Monitoring Daemon
After=network.target

[Service]
Type=simple
ExecStart=%s daemon
Restart=always
RestartSec=5
Environment=MOON_CONFIG=%s

[Install]
WantedBy=default.target
`, binPath(), cfgPath())

	if err := os.WriteFile(svcPath(), []byte(content), 0644); err != nil {
		log.Fatalf("write service: %v", err)
	}
	fmt.Println("service:", svcPath())

	exec.Command("systemctl", "--user", "daemon-reload").Run()
	exec.Command("systemctl", "--user", "enable", "--now", "moon").Run()
	fmt.Println("service enabled and started")

	fmt.Println()
	fmt.Println("install complete")
	fmt.Println()
	fmt.Println("To keep the service running after logout:")
	fmt.Println("  sudo loginctl enable-linger $(whoami)")
}

func cmdUninstall() {
	fmt.Print("stopping and disabling service... ")
	exec.Command("systemctl", "--user", "disable", "--now", "moon").Run()
	if err := os.Remove(svcPath()); err == nil {
		fmt.Println("done")
	} else {
		fmt.Println("not installed")
	}
	exec.Command("systemctl", "--user", "daemon-reload").Run()

	fmt.Print("removing binary... ")
	if err := os.Remove(binPath()); err != nil {
		fmt.Println("not found")
	} else {
		fmt.Println("done")
	}

	fmt.Print("removing config... ")
	if err := os.RemoveAll(filepath.Join(xdgConfigDir(), "moon")); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("done")
	}

	fmt.Print("removing data... ")
	if err := os.RemoveAll(filepath.Join(xdgDataDir(), "moon")); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("done")
	}

	fmt.Println("uninstall complete")
}

func cmdUpdate() {
	wasRunning := false
	out, _ := exec.Command("systemctl", "--user", "is-active", "moon").Output()
	if strings.TrimSpace(string(out)) == "active" {
		wasRunning = true
		exec.Command("systemctl", "--user", "stop", "moon").Run()
	}

	cfg, err := config.Load(cfgPath())
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	repo := cfg.UpdateRepo
	if repo == "" {
		repo = "Alice088/moon"
	}

	fmt.Printf("current version: %s\n", version)
	fmt.Printf("checking %s...\n", repo)

	release, err := fetchLatestRelease(repo)
	if err != nil {
		log.Fatalf("fetch release: %v", err)
	}

	newVer := strings.TrimPrefix(release.TagName, "v")
	curVer := strings.TrimPrefix(version, "v")

	if newVer == curVer || version == "dev" {
		fmt.Println("already up to date")
		if wasRunning {
			exec.Command("systemctl", "--user", "start", "moon").Run()
		}
		return
	}

	fmt.Printf("new version: %s\n", release.TagName)
	if release.Name != "" {
		fmt.Printf("title: %s\n", release.Name)
	}

	downloadURL := ""
	for _, a := range release.Assets {
		if strings.Contains(a.Name, "linux-amd64") && strings.HasSuffix(a.Name, ".tar.gz") {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		log.Fatalf("no linux-amd64 tarball found in release")
	}

	tmpDir, err := os.MkdirTemp("", "moon-update")
	if err != nil {
		log.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tarball := filepath.Join(tmpDir, "release.tar.gz")
	fmt.Print("downloading... ")
	if err := downloadFile(downloadURL, tarball); err != nil {
		log.Fatalf("download: %v", err)
	}
	fmt.Println("done")

	fmt.Print("extracting... ")
	if err := exec.Command("tar", "xzf", tarball, "-C", tmpDir).Run(); err != nil {
		log.Fatalf("extract: %v", err)
	}
	fmt.Println("done")

	newBin := filepath.Join(tmpDir, "moon", "moon")
	if _, err := os.Stat(newBin); err != nil {
		log.Fatalf("binary not found in archive")
	}

	fmt.Print("installing... ")
	data, err := os.ReadFile(newBin)
	if err != nil {
		log.Fatalf("read new binary: %v", err)
	}
	if err := os.WriteFile(binPath(), data, 0755); err != nil {
		log.Fatalf("write binary: %v", err)
	}
	fmt.Println("done")

	fmt.Println("update complete")

	if wasRunning {
		exec.Command("systemctl", "--user", "start", "moon").Run()
	}
}

// --- GitHub release helpers ---

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Name    string    `json:"name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func fetchLatestRelease(repo string) (*ghRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api returned %d: %s", resp.StatusCode, string(body))
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &rel, nil
}

func downloadFile(url, dst string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func cmdVersion() {
	fmt.Printf("moon %s\n", version)

	repo := "Alice088/moon"
	if cfg, err := config.Load(cfgPath()); err == nil && cfg.UpdateRepo != "" {
		repo = cfg.UpdateRepo
	}

	release, err := fetchLatestRelease(repo)
	if err != nil {
		return
	}
	if release.Name != "" {
		fmt.Printf("latest: %s\n", release.Name)
	}
}
