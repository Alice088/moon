package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
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

func main() {
	if runtime.GOOS != "linux" {
		log.Fatalf("only linux supported")
	}

	if os.Geteuid() != 0 {
		log.Fatalf("run as root")
	}

	cfgDir := homeDir() + "/.moon"

	displayArt()

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

	installBinary()
	installConfig(cfgDir)
	installStatic()
	installService(cfgDir)

	fmt.Println()
	fmt.Println("install complete")
	fmt.Println("usage: moon daemon start")
	fmt.Println("       moon daemon enable")
	fmt.Println("       moon status")
}

func displayArt() {
	data, err := os.ReadFile(artFile)
	if err == nil {
		fmt.Println(string(data))
	}

	f, err := os.Open(staticDir + "/" + artFile)
	if err == nil {
		io.Copy(os.Stdout, f)
		f.Close()
		fmt.Println()
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

func installBinary() {
	fmt.Print("installing binary... ")

	src, err := os.Executable()
	if err != nil {
		log.Fatalf("executable path: %v", err)
	}

	data, err := os.ReadFile(src)
	if err != nil {
		log.Fatalf("read self: %v", err)
	}

	dst := installDir + "/" + binaryName
	if err := os.WriteFile(dst, data, 0755); err != nil {
		log.Fatalf("write binary: %v", err)
	}

	fmt.Println("done")
}

func installConfig(cfgDir string) {
	fmt.Print("installing config... ")

	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		log.Fatalf("create config dir: %v", err)
	}

	src := "config.example.yaml"
	if _, err := os.Stat(src); err != nil {
		src = ""
	}

	var data []byte
	if src != "" {
		data, _ = os.ReadFile(src)
	}
	if len(data) == 0 {
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

func installStatic() {
	fmt.Print("installing static files... ")

	if err := os.MkdirAll(staticDir, 0755); err != nil {
		log.Fatalf("create static dir: %v", err)
	}

	if _, err := os.Stat(staticDir + "/" + artFile); err == nil {
		fmt.Println("exists, skip")
		return
	}

	data, err := os.ReadFile("static/" + artFile)
	if err != nil {
		log.Fatalf("read art file: %v", err)
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
ExecStart=%s daemon start
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
