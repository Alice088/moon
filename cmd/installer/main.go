package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
)

const (
	binaryName   = "moon"
	installDir   = "/usr/local/bin"
	configDir    = "/etc/moon"
	staticDir    = "/usr/share/moon/static"
	artFile      = "moon.txt"
)

func main() {
	if runtime.GOOS != "linux" {
		log.Fatalf("only linux supported")
	}

	if os.Geteuid() != 0 {
		log.Fatalf("run as root")
	}

	displayArt()

	installed := checkInstalled()

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
	installConfig()
	installStatic()
	installService()

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

func checkInstalled() bool {
	if _, err := os.Stat(installDir + "/" + binaryName); err == nil {
		return true
	}
	if _, err := os.Stat(configDir + "/config.yaml"); err == nil {
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

func installConfig() {
	fmt.Print("installing config... ")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Fatalf("create config dir: %v", err)
	}

	src := "config.example.yaml"
	if _, err := os.Stat(src); err != nil {
		// embedded or copied alongside
	}

	data, err := os.ReadFile(src)
	if err != nil {
		// minimal default
		data = []byte("storage:\n  db_path: \"/var/lib/moon/moon.db\"\n")
	}

	dst := configDir + "/config.yaml"
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

	data := []byte(`  __  __                   
 |  \/  | ___   _ __   ___ 
 | |\/| |/ _ \ | '_ \ / __|
 | |  | | (_) || | | |\__ \
 |_|  |_|\___/ |_| |_||___/
`)

	if err := os.WriteFile(staticDir+"/"+artFile, data, 0644); err != nil {
		log.Fatalf("write art: %v", err)
	}

	fmt.Println("done")
}

func installService() {
	fmt.Print("installing systemd service... ")

	svc := fmt.Sprintf(`[Unit]
Description=Moon Monitoring Daemon
After=network.target

[Service]
ExecStart=%s daemon start
Restart=always
RestartSec=5
Environment=MOON_CONFIG=/etc/moon/config.yaml

[Install]
WantedBy=multi-user.target
`, installDir+"/"+binaryName)

	path := "/etc/systemd/system/moon.service"
	if err := os.WriteFile(path, []byte(svc), 0644); err != nil {
		log.Fatalf("write service: %v", err)
	}

	exec.Command("systemctl", "daemon-reload").Run()
	exec.Command("systemctl", "enable", "moon").Run()

	fmt.Println("done")
}

