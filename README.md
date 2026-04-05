# lif

Gamified TUI task manager and command reference. Daily habits with streak tracking, rolling todos, countdown reminders, and a searchable command glossary. "Lucas is Forgetful." Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Install

```bash
go install github.com/LFroesch/lif@latest
```

Or use the install script:

```bash
curl -sSL https://raw.githubusercontent.com/LFroesch/lif/main/install.sh | sh
```

Or build from source:

```bash
make install
```

## Usage

```bash
lif
```

## Tabs

### 1. Home
Dashboard with daily activity summary and stats.

### 2. Daily Tasks
Recurring tasks that reset at 3 AM. Build streaks by completing them daily.

| Key | Action |
|-----|--------|
| `space/enter` | Toggle completion |
| `n/a` | Add task |
| `e` | Edit task |
| `d` | Delete task |

Each task tracks current streak and best streak.

### 3. Rolling Todos
Persistent todos that don't reset. Priority-based sorting, category grouping, deadline tracking.

### 4. Reminders
Countdown timers and alarms with system notifications.

| Key | Action |
|-----|--------|
| `s` | Start/resume |
| `p` | Pause |
| `r` | Reset |

**Time formats:** `30s`, `5m`, `2h`, `1d`, `1w` (countdown) or `9:30AM`, `15:30` (alarm).

### 5. Reference
Searchable command glossary with 50+ pre-populated commands (git, docker, npm, curl, bash, Go).

| Key | Action |
|-----|--------|
| `/` | Search |
| `esc` | Clear search |
| `s` | Sort |

## Global Keybindings

| Key | Action |
|-----|--------|
| `1-5` | Switch tabs |
| `j/k`, `up/down` | Navigate |
| `n/a` | Add item |
| `e` | Edit |
| `d` | Delete (with confirmation) |
| `?` | Help |
| `q` | Quit |

## Configuration

All data saved to `~/.config/lif/config.json`. No external database.

Priority system: HIGH (red), MEDIUM (yellow), LOW (green).

## Platform Support

Linux, macOS, WSL. Notifications: native on Linux/macOS, PowerShell toast on WSL.

## License

[AGPL-3.0](LICENSE)
