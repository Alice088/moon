package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
)

const (
	binaryName = "moon"
	installDir = "/usr/local/bin"
	staticDir  = "/usr/share/moon/static"
	artFile    = "art.txt"
)

func homeDir() string {
	u, err := user.Current()
	if err != nil {
		return "/root"
	}
	return u.HomeDir
}

func selfDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func main() {
	if runtime.GOOS != "linux" {
		log.Fatalf("only linux supported")
	}
	if os.Geteuid() != 0 {
		log.Fatalf("run as root")
	}

	dir := selfDir()
	cfgDir := homeDir() + "/.moon"

	displayArt(dir)

	installed := checkInstalled(cfgDir)
	if installed {
		fmt.Println("moon already installed")
		fmt.Println("reinstall? [y/N]")
		var reply string
		fmt.Scanln(&reply)
		if reply != "y" && reply != "Y" {
			fmt.Println("cancelled")
			os.Exit(0)
		}
	}

	installBinary(dir)
	installConfig(dir, cfgDir)
	installStatic(dir)
	installService(cfgDir)

	fmt.Println()
	fmt.Println("install complete")
	fmt.Println("usage: moon start")
	fmt.Println("       moon enable")
	fmt.Println("       moon status")
}

func displayArt(dir string) {
	data, err := os.ReadFile(dir + "/static/" + artFile)
	if err == nil {
		fmt.Println(string(data))
	}
}

func checkInstalled(cfgDir string) bool {
	if _, err := os.Stat(installDir + "/" + binaryName); err == nil {
		return true
	}
	if _, err := os.Stat(cfgDir + "/config.yaml"); err == nil {
		return true
	}
	return false
}

func installBinary(dir string) {
	fmt.Print("installing binary... ")

	src := dir + "/" + binaryName
	data, err := os.ReadFile(src)
	if err != nil {
		log.Fatalf("read moon binary: %v", err)
	}

	dst := installDir + "/" + binaryName
	if err := os.WriteFile(dst, data, 0755); err != nil {
		log.Fatalf("write binary: %v", err)
	}

	fmt.Println("done")
}

func installConfig(dir, cfgDir string) {
	fmt.Print("installing config... ")

	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		log.Fatalf("create config dir: %v", err)
	}

	src := dir + "/config.example.yaml"
	data, err := os.ReadFile(src)
	if err != nil {
		data = []byte("storage:\n  db_path: \"" + cfgDir + "/moon.db\"\n")
	}

	dst := cfgDir + "/config.yaml"
	if _, err := os.Stat(dst); err == nil {
		fmt.Println("exists, skip")
		return
	}

	if err := os.WriteFile(dst, data, 0644); err != nil {
		log.Fatalf("write config: %v", err)
	}

	fmt.Println("done")
}

func installStatic(dir string) {
	fmt.Print("installing static files... ")

	if err := os.MkdirAll(staticDir, 0755); err != nil {
		log.Fatalf("create static dir: %v", err)
	}

	if _, err := os.Stat(staticDir + "/" + artFile); err == nil {
		fmt.Println("exists, skip")
		return
	}

	data, err := os.ReadFile(dir + "/static/" + artFile)
	if err != nil {
		log.Fatalf("read art: %v", err)
	}

	if err := os.WriteFile(staticDir+"/"+artFile, data, 0644); err != nil {
		log.Fatalf("write art: %v", err)
	}

	fmt.Println("done")
}

func installService(cfgDir string) {
	fmt.Print("installing systemd service... ")

	svc := fmt.Sprintf(`[Unit]
Description=Moon Monitoring Daemon
After=network.target

[Service]
ExecStart=%s start
Restart=always
RestartSec=5
Environment=MOON_CONFIG=%s/config.yaml

[Install]
WantedBy=multi-user.target
`, installDir+"/"+binaryName, cfgDir)

	path := "/etc/systemd/system/moon.service"
	if err := os.WriteFile(path, []byte(svc), 0644); err != nil {
		log.Fatalf("write service: %v", err)
	}

	exec.Command("systemctl", "daemon-reload").Run()
	exec.Command("systemctl", "enable", "moon").Run()

	fmt.Println("done")
}
