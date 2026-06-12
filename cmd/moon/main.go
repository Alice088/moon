package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"moon/internal/daemon"
)

const (
	pidFile = "/var/run/moon.pid"
	svcFile = "/etc/systemd/system/moon.service"
)

func homeDir() string {
	u, err := user.Current()
	if err != nil {
		return "/root"
	}
	return u.HomeDir
}

func defaultCfgPath() string {
	if v := os.Getenv("MOON_CONFIG"); v != "" {
		return v
	}
	return homeDir() + "/.moon/config.yaml"
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "daemon":
		if len(os.Args) < 3 {
			printUsage()
			os.Exit(1)
		}
		switch os.Args[2] {
		case "start":
			cmdDaemonStart()
		case "stop":
			cmdDaemonStop()
		case "enable":
			cmdDaemonEnable()
		case "disable":
			cmdDaemonDisable()
		default:
			printUsage()
			os.Exit(1)
		}

	case "status":
		cmdStatus()

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`moon — system monitoring daemon

Commands:
  daemon start     start monitoring daemon
  daemon stop      stop monitoring daemon
  daemon enable    install systemd service (autostart on boot)
  daemon disable   remove systemd service
  status           show daemon status`)
}

func cmdDaemonStart() {
	if pidRunning() {
		log.Println("already running")
		os.Exit(1)
	}

	if err := daemon.Run(defaultCfgPath()); err != nil {
		log.Fatalf("daemon error: %v", err)
	}
}

func cmdDaemonStop() {
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

func cmdDaemonEnable() {
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
ExecStart=%s daemon start
Restart=always
RestartSec=5
Environment=MOON_CONFIG=%s

[Install]
WantedBy=multi-user.target
`, exe, homeDir()+"/.moon/config.yaml")

	if err := os.WriteFile(svcFile, []byte(content), 0644); err != nil {
		log.Fatalf("write service: %v", err)
	}

	exec.Command("systemctl", "daemon-reload").Run()
	exec.Command("systemctl", "enable", "moon").Run()

	log.Println("service installed and enabled")
}

func cmdDaemonDisable() {
	exec.Command("systemctl", "disable", "moon").Run()
	os.Remove(svcFile)
	exec.Command("systemctl", "daemon-reload").Run()
	log.Println("service disabled and removed")
}

func cmdStatus() {
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
