# moon

Minimal, zero-dependency server monitoring daemon. Collects CPU, RAM, and disk metrics, detects peak loads, stores history in SQLite, and sends alerts to Telegram.

Designed for headless servers — single static binary, no runtime deps (no Python, no Node, no CGO), reads directly from `/proc`. Runs as a systemd user service with automatic restarts.

## Features

- CPU, RAM, disk usage collectors (reads from `/proc` and `syscall.Statfs`)
- Peak detection analyzers with configurable thresholds
- SQLite storage for alert history
- Telegram bot — query peaks and averages with arbitrary periods
- Interval-based reporting — divides any period into 3 parts, shows peak + time
- Modular architecture — swap Telegram for any channel (email, Slack, webhook) via pluggable notifiers
- Extensible — pluggable collectors, analyzers, hooks, and notifiers
- Systemd integration for autostart

## Install

Download the latest release and run:

```bash
tar xzf moon-*.tar.gz
./moon install
```

Or build from source and install:

```bash
make build
build/moon install
```

Installer copies the binary to `~/.local/bin/moon`, creates config in `~/.config/moon/config.yaml`, and installs a systemd user service with auto-restart.

## Quick start

```bash
moon start     # start via systemd --user
moon stop      # stop via systemd --user
moon status    # show version and running status
moon daemon    # run in foreground (used by systemd)
moon update    # update to latest release
moon version   # print version
moon uninstall # stop, remove binary, config, data, service
```

> ⚠️ Run `sudo loginctl enable-linger $(whoami)` to keep the service running after logout.

## Telegram bot

Periods are flexible — any number + suffix works: `10m`, `56m`, `5h`, `3h`, `1d`, `89d`, `365d`, `2w`, `6mo`. The bot divides the period into 3 equal intervals and reports per interval.

```
/peaks <period>     — 3 intervals, highest peak + time per interval
/peak-avg <period>  — 3 intervals, average per interval
```

### Examples

```
/peaks 1d
Peaks Jun 12 — Jun 13:
  cpu:
    Jun 12 16:48: 91.9%
    Jun 13 04:31: 89.8%
    Jun 13 13:36: 98.7%
  ram:
    Jun 12 17:01: 94.3%
    Jun 12 22:40: 81.3%
  disk:
    Jun 12 22:48: 97.4%
    Jun 13 11:28: 94.6%

/peaks 1h
Peaks 13:24 — 14:24:
  cpu:
    13:36: 98.7%
    13:50: 97.7%
    14:14: 98.7%

/peaks 365d
Peaks Jun 13 — Jun 13:
  cpu:
    Dec 28 20:56: 98.2%
    Jun 12 09:25: 99.9%
  ram:
    Jan 31 02:11: 91.1%
    May 25 03:19: 100.0%
  disk:
    Jan 11 10:14: 82.2%
    May 19 13:45: 99.9%

/peak-avg 365d
Avg peaks Jun 13 — Jun 13:
  cpu:
    Feb 11 22:25: 86.0%
    Jun 13 14:25: 87.0%
  ram:
    Feb 11 22:25: 82.5%
    Jun 13 14:25: 84.1%
  disk:
    Feb 11 22:25: 77.8%
    Jun 13 14:25: 85.0%
```


## Configuration

Config file: `~/.config/moon/config.yaml` (override with `MOON_CONFIG` env var).

```yaml
cpu:
  peak_threshold_pct: 80.0

ram:
  peak_threshold_pct: 80.0

disk:
  peak_threshold_pct: 80.0

analyzer_workers: 2
hook_workers: 2
collect_interval: 1s

update_repo: "Alice088/moon"
debug: false

storage:
  db_path: "~/.local/share/moon/moon.db"

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
Pipeline (collectors @ 1s interval)
  |
  v
AnalyzerPool (worker pool, detects threshold breaches)
  |
  +-> HookRunner → SQLite (WriteAlertToDB)
  |
  +-> Dispatcher → Notifiers (Telegram)

Bot (long poll Telegram API)
  +-> storage.PeakByIntervals()  — 3 intervals, MAX + time
  +-> storage.AverageByIntervals() — 3 intervals, AVG
```

- **Collectors** read `/proc/stat`, `/proc/meminfo`, `syscall.Statfs("/")`
- **Analyzers** compare against thresholds, create alerts
- **Hooks** persist alerts to SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- **Dispatcher** fans out alerts to all notifiers
- **Bot** polls Telegram, responds with interval-based peak/avg queries

## Project structure

```
cmd/
  moon/        CLI (daemon, install, uninstall, update, status)
internal/
  analyzer/    Peak detectors (CPU, RAM, Disk)
  bot/         Telegram bot (/peaks, /peak-avg commands)
  collector/   /proc metrics readers
  config/      YAML config loader (~ expansion, embed defaults)
  daemon/      Pipeline wiring (collect → analyze → hook → dispatch)
  dispatcher/  Alert fan-out to notifiers
  entity/      Types (Metrics, Alert, Pipeline)
  hook/        SQLite writer hook
  notifier/    Telegram sender
  storage/     SQLite queries (PeakByIntervals, AverageByIntervals)
```
