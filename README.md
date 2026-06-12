# moon

System monitoring daemon. Collects CPU, RAM, and disk metrics, detects peak loads, and sends alerts to Telegram.

## Features

- CPU, RAM, disk usage collectors (reads from `/proc` and `syscall.Statfs`)
- Peak detection analyzers with configurable thresholds
- SQLite storage for alert history
- Telegram notifications (no emoji, Markdown formatted)
- Plugable hook system for custom output
- Systemd integration for autostart
- Single binary, no runtime dependencies

## Install

Download the latest release from GitHub and run:

```bash
sudo tar xzf moon-*.tar.gz -C /tmp
sudo /tmp/moon/moon-installer
```

Installer copies the binary to `/usr/local/bin`, creates config in `/root/.moon/config.yaml`, installs systemd service, and deploys static assets.

## Quick start

```bash
moon enable    # install and enable systemd service
moon start     # start the daemon
moon status    # check if running
moon stop      # stop the daemon
moon disable   # remove systemd service
```

## Configuration

Config file: `/root/.moon/config.yaml` (override with `MOON_CONFIG` env var).

```yaml
cpu:
  peak_threshold_pct: 80.0

ram:
  peak_threshold_pct: 80.0

disk:
  peak_threshold_pct: 80.0

analyzer_workers: 2
hook_workers: 2

storage:
  db_path: "/root/.moon/moon.db"

notify:
  - type: telegram
    bot_token: "123456:ABC-DEF"
    chat_id: "-1001234567890"
```

## Build from source

```bash
git clone <repo> && cd moon
make build       # binaries in build/
make dist        # release tarball in build/release/
make clean       # remove build/
```

Requires Go 1.26.3. Dependencies vendored with Go modules (pure Go SQLite via `modernc.org/sqlite`, no CGO).

## Architecture

```
Pipeline (collectors)
  |
  v
AnalyzerPool (worker pool)
  |
  +-> HookRunner (SQLite writer, etc.)
  |
  +-> Dispatcher -> Notifiers (Telegram, etc.)
```

- **Collectors** read `/proc/stat`, `/proc/meminfo`, and `syscall.Statfs("/")`
- **Analyzers** detect peaks and write alerts into shared Metrics snapshot
- **Hooks** store alerts to SQLite via `WriteAlertToDB`
- **Dispatcher** picks up alerts and sends via configured notifiers
- **Peak queries** use SQL `MAX`/`AVG` with `json_extract` on stored alert data

## Project structure

```
cmd/
  moon/          CLI binary (daemon start/stop/enable/disable, status)
  installer/     Setup tool (deploys binary, config, systemd unit)
internal/
  analyzer/      Peak detectors (CPU, RAM, Disk)
  collector/     System metrics readers
  config/        YAML config loader
  daemon/        Pipeline wiring
  dispatcher/    Alert fan-out to notifiers
  entity/        Types (Metrics, Alert, Pipeline, Analyzer, Collector)
  hook/          Hook interface, runner, write_alert_to_db
  notifier/      Telegram sender
  storage/       SQLite queries (Peak, Average, ListAlerts)
static/
  art.txt        ASCII banner
```
