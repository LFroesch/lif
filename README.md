# lif - Lucas is Forgetful

A gamified terminal user interface (TUI) application for managing daily tasks, reminders, rolling todos, and a command glossary. Perfect for forgetful minds who want to build consistent habits through daily tasks while earning points, unlocking achievements, and maintaining streaks!

## Features

### 🎮 Gamification System
- **Points & Levels**: Earn points and level up!
  - Daily tasks: 10 points each
  - Daily login bonus: 5 points
  - Level up every 100 points!
- **Streaks**: Track your daily login streak
- **Achievements**: Unlock 11 different achievements
  - First Steps, Consistent Checker (7-day login)
  - On Fire! (3-day task streak), Week Warrior (7-day)
  - Fortnight Force (14-day), Monthly Master (30-day)
  - Taskmaster (10 tasks), Productivity Pro (50), Century Club (100)
  - Level 5 Hero, Elite Achiever (Level 10)
- **Progress Dashboard**: See your stats, level, points, and streaks at a glance

### 📋 Daily Tasks (Your Habits!)
- Create recurring daily tasks that reset at 3 AM
- Build consistency with **streak tracking** - complete tasks daily to build streaks!
- Track current streak and best streak for each task
- Visual streak indicators with 🔥
- Organize by priority (HIGH/MEDIUM/LOW) and category
- Set deadlines and monitor progress
- Earn 10 points for completing tasks!

### 🔄 Rolling Todos
- Persistent todo items that don't reset daily
- Priority-based organization
- Category grouping for better organization
- Deadline tracking

### ⏰ Reminders & Alarms
- Set countdown timers (1m, 30s, 2h, 5d, etc.)
- Schedule alarms for specific times (9:30AM, 15:30)
- Pause and resume countdowns
- System notifications with sound alerts
- Cross-platform notification support (Linux, macOS, Windows, WSL)

### 📚 Command Reference
- **Pre-populated with 50+ common commands** for git, docker, npm, curl, bash, and Go
- **Live search functionality** - press `/` to quickly find commands
- Store your own custom commands and snippets
- Organize by programming language or category
- Quick reference with usage examples and meanings
- Perfect for forgetful devs who need a quick `man` page alternative

## Installation

```bash
go install github.com/LFroesch/lif@latest
```

Make sure `$GOPATH/bin` (usually `~/go/bin`) is in your PATH:
```bash
export PATH="$HOME/go/bin:$PATH"
```

## Usage

### Navigation
- **Numbers 1-5**: Switch between tabs
  - 1: Home (Dashboard with gamification stats & streaks)
  - 2: Daily Tasks (with streak tracking!)
  - 3: Rolling Todos
  - 4: Reminders
  - 5: Reference (Command lookup with search)
- **Left/Right arrows**: Navigate tabs
- **Up/Down arrows** or **j/k**: Navigate within tables

### Basic Operations
- **e**: Edit selected item
- **n** or **a**: Add new item
- **d**: Delete selected item (with confirmation)
- **q**: Quit application

### Tab-Specific Controls

#### Home (Tab 1)
- View your gamification stats:
  - Current level, total points, and login streak
  - Daily activity summary
  - Tasks completed
  - Unlocked achievements
  - Active task streaks

#### Daily Tasks (Tab 2)
- **Space** or **Enter**: Toggle task completion
- Tasks automatically reset to incomplete at 3 AM daily
- Complete tasks daily to build streaks!
- Earn 10 points per task completed
- View current streak and best streak for each task

#### Reminders (Tab 4)
- **s**: Start/resume reminder
- **p**: Pause active reminder
- **r**: Reset reminder to original time

#### Reference (Tab 5)
- **/**: Activate search mode
- **ESC**: Clear search and return to browsing
- **s**: Sort by language or command
- Search across all fields (language, command, usage, example, meaning)
- Pre-populated with common git, docker, npm, curl, bash, and Go commands

### Time Formats

#### For Countdowns
- **Seconds**: `30s`, `45sec`
- **Minutes**: `5m`, `30min`
- **Hours**: `2h`, `3hr`
- **Days**: `1d`, `7d`
- **Weeks**: `1w`, `2w`

#### For Alarms
- **12-hour format**: `9:30AM`, `2:15 PM`
- **24-hour format**: `09:30`, `14:15`

## Configuration

Configuration is automatically saved to:
-  `~/.config/lif/config.json`

## Features in Detail

### Smart Notifications
- Audio alerts with fallback to system beep
- Supports multiple audio formats (MP3, WAV)
- WSL-compatible notification system

### Priority System
- **HIGH**: Red styling, highest priority in sorting
- **MEDIUM**: Yellow styling, default priority
- **LOW**: Green styling, lowest priority

### Data Persistence
- All data automatically saved to JSON configuration
- No external database required
- Portable configuration file

### Visual Design
- Color-coded priority indicators
- Modern table styling with clean borders
- Status-aware color themes
- Responsive layout that adapts to terminal size

## Keyboard Shortcuts Reference

| Key | Action | Context |
|-----|--------|---------|
| `1-5` | Switch tabs | Global |
| `←/→` | Navigate tabs | Global |
| `↑/↓` or `j/k` | Navigate items | Tables |
| `e` | Edit selected | Tables |
| `n/a` | Add new item | Tables |
| `d` | Delete item | Tables |
| `Space/Enter` | Toggle completion (builds streaks!) | Daily Tasks |
| `s` | Start/resume / Sort | Reminders / Reference |
| `p` | Pause | Reminders |
| `r` | Reset | Reminders |
| `/` | Search | Reference |
| `ESC` | Clear search | Reference (when searching) |
| `q` | Quit | Global |

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling