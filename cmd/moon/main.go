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
	"strconv"
	"strings"
	"syscall"

	"moon/internal/config"
	"moon/internal/daemon"
)

var version = "dev"

func requireRoot() {
	if os.Geteuid() != 0 {
		log.Fatalf("must run as root")
	}
}

const (
	pidFile    = "/var/run/moon.pid"
	svcFile    = "/etc/systemd/system/moon.service"
	cfgPath    = "/root/.moon/config.yaml"
	binPath    = "/usr/local/bin/moon"
	cfgDir     = "/root/.moon"
	staticDir  = "/usr/share/moon/static"
)

func effectiveCfgPath() string {
	if v := os.Getenv("MOON_CONFIG"); v != "" {
		return v
	}
	return cfgPath
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "start":
		cmdStart()
	case "stop":
		cmdStop()
	case "enable":
		cmdEnable()
	case "disable":
		cmdDisable()
	case "status":
		cmdStatus()
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
  start     start monitoring
  stop      stop monitoring
  enable    install systemd service (autostart on boot)
  disable   remove systemd service
  status    show daemon status
  uninstall remove all files (binary, config, service)
  update    update to latest version
  version   print version`)
}

func cmdStart() {
	requireRoot()
	if pidRunning() {
		log.Println("already running")
		os.Exit(1)
	}

	if os.Getenv("_MOON_FG") == "" {
		daemonize()
		return
	}

	writePID()
	defer os.Remove(pidFile)

	if err := daemon.Run(effectiveCfgPath()); err != nil {
		log.Fatalf("daemon error: %v", err)
	}
}

func daemonize() {
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("executable path: %v", err)
	}

	attr := &os.ProcAttr{
		Files: []*os.File{nil, nil, nil},
		Env:   append(os.Environ(), "_MOON_FG=1"),
	}

	proc, err := os.StartProcess(exe, []string{exe, "start"}, attr)
	if err != nil {
		log.Fatalf("fork: %v", err)
	}
	log.Printf("started (pid %d)", proc.Pid)
	os.Exit(0)
}

func writePID() {
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())+"\n"), 0644); err != nil {
		log.Fatalf("write pid: %v", err)
	}
}

func cmdStop() {
	requireRoot()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		log.Fatalf("not running (no pid file)")
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		log.Fatalf("invalid pid: %v", err)
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		log.Fatalf("find process: %v", err)
	}

	if err := p.Signal(syscall.SIGTERM); err != nil {
		log.Fatalf("stop: %v", err)
	}

	os.Remove(pidFile)
	log.Println("stopped")
}

func cmdEnable() {
	requireRoot()
	if _, err := os.Stat(svcFile); err == nil {
		log.Println("service already installed")
		os.Exit(1)
	}

	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("executable path: %v", err)
	}

	content := fmt.Sprintf(`[Unit]
Description=Moon Monitoring Daemon
After=network.target

[Service]
Type=forking
PIDFile=%s
ExecStart=%s start
Restart=always
RestartSec=5
Environment=MOON_CONFIG=%s

[Install]
WantedBy=multi-user.target
`, pidFile, exe, cfgPath)

	if err := os.WriteFile(svcFile, []byte(content), 0644); err != nil {
		log.Fatalf("write service: %v", err)
	}

	exec.Command("systemctl", "daemon-reload").Run()
	exec.Command("systemctl", "enable", "moon").Run()

	log.Println("service installed and enabled")
}

func cmdDisable() {
	requireRoot()
	exec.Command("systemctl", "disable", "moon").Run()
	os.Remove(svcFile)
	exec.Command("systemctl", "daemon-reload").Run()
	log.Println("service disabled and removed")
}

func cmdStatus() {
	if version != "" {
		fmt.Printf("version: %s\n", version)
	}

	if pidRunning() {
		data, _ := os.ReadFile(pidFile)
		fmt.Printf("running (pid %s)\n", strings.TrimSpace(string(data)))
	} else {
		fmt.Println("not running")
	}

	if _, err := os.Stat(svcFile); err == nil {
		fmt.Println("autostart: enabled")
	} else {
		fmt.Println("autostart: disabled")
	}
}

func cmdUpdate() {
	requireRoot()

	wasRunning := pidRunning()
	if wasRunning {
		cmdStop()
	}

	cfg, err := config.Load(effectiveCfgPath())
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
			cmdStart()
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
	os.Remove(binPath)
	if err := os.WriteFile(binPath, data, 0755); err != nil {
		log.Fatalf("write binary: %v", err)
	}
	fmt.Println("done")

	fmt.Println("update complete")

	if wasRunning {
		cmdStart()
	}
}

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

func cmdUninstall() {
	requireRoot()

	fmt.Print("stopping daemon... ")
	if pidRunning() {
		cmdStop()
	} else {
		fmt.Println("not running")
	}

	fmt.Print("disabling service... ")
	if _, err := os.Stat(svcFile); err == nil {
		exec.Command("systemctl", "disable", "moon").Run()
		os.Remove(svcFile)
		exec.Command("systemctl", "daemon-reload").Run()
		fmt.Println("done")
	} else {
		fmt.Println("not installed")
	}

	fmt.Print("removing binary... ")
	if err := os.Remove(binPath); err != nil {
		fmt.Println("not found")
	} else {
		fmt.Println("done")
	}

	fmt.Print("removing config... ")
	if err := os.RemoveAll(cfgDir); err != nil {
		fmt.Println("not found")
	} else {
		fmt.Println("done")
	}

	fmt.Print("removing static files... ")
	if err := os.RemoveAll(staticDir); err != nil {
		fmt.Println("not found")
	} else {
		fmt.Println("done")
	}

	fmt.Println("uninstall complete")
}

func cmdVersion() {
	fmt.Printf("moon %s\n", version)

	repo := "Alice088/moon"
	if cfg, err := config.Load(effectiveCfgPath()); err == nil && cfg.UpdateRepo != "" {
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

func pidRunning() bool {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}
