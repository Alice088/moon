# v0.2.0 — Interval peaks, timezone fix, config improvements

## Highlights

- **`/peaks` and `/peak-avg` now show 3 time intervals** instead of a single aggregate value. Each interval shows the peak (with exact time) or average, making it easy to spot when spikes happened.
- **Timezone bug fixed** — SQLite stores UTC, but queries used local time. In MSK (UTC+3), this caused all comparisons to miss by 3 hours, returning `0.0%` in Telegram.
- **Config generation fixed** — embedded `config.example.yaml` directly into the binary so `moon install` always writes a complete config, even when running via `go run`.

## Features

### Interval-based peaks (`/peaks`)
Divides the requested period into 3 equal parts and shows the highest peak with the exact time it occurred:

```
Peaks 13:01 — 14:01:
  cpu:
    13:21: no data
    13:41: no data
    13:56: 92.3%
  ram:
    13:21: no data
    13:41: no data
    13:58: 76.0%
  disk:
    13:21: no data
    13:41: no data
    14:01: 91.0%
```

For long periods (≥24h), timestamps show date + time (`Jun 13 15:04`).

### Interval-based averages (`/peak-avg`)
Same 3-interval split, shows average usage per interval. Intervals with no alerts show `0.0%`:

```
Avg peaks 13:01 — 14:01:
  cpu:
    13:21: 0.0%
    13:41: 0.0%
    14:01: 88.7%
  ram:
    13:21: 0.0%
    13:41: 0.0%
    14:01: 76.0%
  disk:
    13:21: 0.0%
    13:41: 0.0%
    14:01: 91.0%
```

### Debug mode
Added `debug: true` to `config.yaml`. When enabled, logs every storage query (SQL, timestamps, results) and bot message handling. Useful for diagnosing why data isn't showing up.

## Bug Fixes

### Timezone mismatch → 0.0% in Telegram
- **Root cause:** `storage.Peak()` and `storage.Average()` formatted `since` in local time (`since.Format(...)`). SQLite stores `created_at` via `CURRENT_TIMESTAMP` which is always UTC. In MSK (UTC+3), `since` formatted as `13:41:00` instead of `10:41:00` UTC → `WHERE created_at >= '13:41:00'` matched nothing → `COALESCE(..., 0)` → `0.0%`.
- **Fix:** Use `since.UTC().Format(...)` in all storage queries so both sides of the comparison are in UTC.

### `~` in `db_path` not expanded
- **Root cause:** `config.example.yaml` had `db_path: "~/.local/share/moon/moon.db"`. SQLite receives the literal string `~/.local/share/...` and can't open it, returning error 14.
- **Fix:** `config.Load()` now detects `~/` prefix and expands it via `os.UserHomeDir()`.

### Database directory not created
- **Root cause:** Running `go run ./cmd/moon/main.go daemon` without `install` skipped the data directory creation. SQLite couldn't create the `.db` file in a non-existent parent directory.
- **Fix:** `daemon.Run()` now calls `os.MkdirAll()` on the DB directory before any database operations.

### Incomplete config on `go run install`
- **Root cause:** The installer looked for `config.example.yaml` next to the running binary (`filepath.Dir(exe)`). When running via `go run`, the binary is in a temp directory, the file isn't found, and the fallback wrote only `storage:\n  db_path: ...` — missing all other fields (cpu/ram/disk thresholds, notify config, etc.).
- **Fix:** Used `//go:embed` to embed `config.example.yaml` into the binary. The installer always writes the complete config, replacing `~/.local/share/moon` with the expanded home path.

## Technical Changes

| File | Change |
|------|--------|
| `cmd/moon/main.go` | Embedded `config.example.yaml` via `//go:embed`; install writes full config instead of one-line stub |
| `cmd/moon/config.example.yaml` | Added `debug: false` field |
| `config.example.yaml` | Added `debug: false` field |
| `internal/config/config.go` | Added `Debug bool` field; expand `~/` in `db_path` |
| `internal/storage/peak.go` | Added `PeakByIntervals()` function; `IntervalPeak` struct; `Debug` var |
| `internal/storage/average.go` | Added `AverageByIntervals()` function; `Debug` var |
| `internal/storage/peak_test.go` | Seed data uses UTC to match production |
| `internal/bot/new_bot.go` | Added `debug` field, `SetDebug()` method |
| `internal/bot/handle_message.go` | Rewrote `sendPeaks` (3 intervals, list format, peak time); rewrote `sendPeakAvg` (3 intervals) |
| `internal/bot/post_json.go` | Mocks `postJSONFunc` for testing |
| `internal/bot/send_peaks_test.go` | Updated all assertions for new output format |
| `internal/daemon/run.go` | Creates DB directory on start; sets `storage.Debug` and `bot.SetDebug` from config |
| `cmd/seed/main.go` | Temporary seed script (created for testing, removed after use) |

## Upgrade

```bash
moon update
```

Or reinstall:

```bash
moon uninstall
go run ./cmd/moon/main.go install
```

> ⚠️ Reinstall overwrites `~/.config/moon/config.yaml`. Back up your config first if you've customized it.
