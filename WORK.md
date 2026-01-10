# lif - Work Log

## Current Tasks

Bugs:
notifications don't work

### Architecture Refactor
- [ ] Split main.go into modular structure (model.go, update.go, view.go, helpers.go)
- [ ] Create internal/ packages for logical grouping
- [ ] Other refactoring / cleanup?

### Distribution Setup
- [ ] Update README with install options (curl, go install, binary downloads)

### UX Improvements
- [ ] Improve dashboard layout and information hierarchy
- [ ] Add modal-style confirmation dialogs
- [ ] Enhance table styling and readability

## Future Tasks
- Better keyboard navigation (vim-style gg/G, Ctrl+D/U)
- make it so you can do "lif todo talk to so and so" and it will add it to your rolling tasks or "lif r 4:00pm laundry" starts an alarm clock reminder at 4 or whatever or lif todo -l or something lists all your todos out or idk some stuff like that 

## DevLog

### 2026-01-09 - Daily Task Streak System Rewrite

**Major Streak Bug Fixes:**
- Fixed core issue: LastCompleted was being set BEFORE updateTaskStreak() was called (update.go:665)
- This meant streak calculation always saw "today" as last completion, never "yesterday"
- Streaks would always reset to 1 instead of incrementing properly
- Moved LastCompleted assignment to AFTER updateTaskStreak() so it can check the previous value

**Simplified Streak Tracking:**
- Removed CompletedDays array field - was redundant and never cleaned up
- Rewrote updateTaskStreak() (gamification.go:17-46) to use only LastCompleted timestamp
- Now uses get3AMDay() to check both "already completed today" and "completed yesterday"
- Much simpler logic without array tracking

**Additional Fix:**
- In resetDailyTasks() (helpers.go:316), added CurrentStreak reset when task status is reset
- Prevents stale streak values from persisting after daily reset

**Config Reset:**
- Reset all daily task streaks to 0 in user config file
- Removed all completed_days arrays from existing data

**Result:**
- Streaks now properly increment when tasks are completed on consecutive days
- Clean slate for testing the fixed streak system

### 2026-01-09 - Help Page Refactor to Match Scout

**Help View Improvements:**
- Made help page scrollable with scroll indicators (▲/▼) like scout
- Changed from centered modal to full-width bordered panel
- Updated formatting to match scout's style:
  - Orange section headers (color 214)
  - Purple key bindings (color 105)
  - Consistent spacing and layout
- Added helpScroll field to model for scroll position tracking
- Implemented scroll navigation (up/down/j/k) in update.go
- Help scroll resets to 0 when closing help screen

**Files Modified:**
- model.go: Added helpScroll field
- view.go: Completely rewrote helpView() with scrollable implementation
- update.go: Added scroll handling (up/down/j/k keys) in help mode

**Result:**
- Help page now matches scout's UX with scrollable content
- Better use of screen space with full-width panel
- Consistent visual style across tui-hub apps

### 2026-01-09 - Home Page Consistency & Status Bar Fix

**Visual Consistency:**
- Removed border wrapper from home page view (view.go:456-458)
- Removed "🏠 Dashboard" header from home page (view.go:342-458)
- Home page now matches styling of Dailies, Rolling, Reminders, and Reference tabs
- Fixed status bar background gaps (helpers.go:45-47)
  - Added .Background(lipgloss.Color("240")) to keyStyle, actionStyle, and bulletStyle
  - Eliminates gaps showing terminal color in status bar
- Converted raw ANSI codes to lipgloss (view.go:39)
  - Changed "\033[1;32m✓ DONE\033[0m" to statusDoneStyle.Render("✓ DONE")
  - Consistent styling approach throughout codebase

**Result:**
- Consistent visual appearance across all tabs
- Status bar has uniform grey background throughout without gaps
- Home page displays only the title "📋 lif - lucas is forgetful" like other pages
- All styling now uses lipgloss properly, no raw ANSI escape codes

### 2026-01-08 - Home Page Layout & 3AM Streak System Fix

**Home Page Simplification:**
- Removed "Reference Entries" stat from home dashboard
- Removed "Active Task Streaks" section entirely
- Home page now shows only essential stats: Daily Tasks, Rolling Todos, Active Reminders

**Streak System 3AM Cutoff Implementation:**
- Added get3AMDay() helper function in gamification.go
- Updated updateTaskStreak() to use 3AM-based day calculation instead of calendar days
- Completions before 3AM now count for the previous day (e.g., 2AM = yesterday, 4AM = today)
- Updated resetDailyTasks() in helpers.go to break streaks when tasks are skipped
- Current streak and best streak are now properly separated (best tracks all-time record)

**Result:**
- Home page is cleaner and matches other pages better
- Streak system correctly handles the 3AM cutoff window
- Streaks properly reset when tasks are skipped for a day

### 2026-01-08 - View Refactor to Match Scout's Panel System

**UI Architecture Improvements:**
- Refactored View() to match scout's panel-based system with separate render functions
- Extracted renderHeader() and renderStatusBar() methods for better separation of concerns
- Added terminal size detection with helpful warning message (like scout)
- Implemented switch-based content rendering (cleaner than if-else chains)
- Added dimension constants (minTerminalWidth, minTerminalHeight, uiOverhead)
- Added helper methods: getSafeWidth(), getSafeHeight(), getContentHeight()
- Moved confirmDelete dialog to renderConfirmDeleteView() for consistency

**Home View Fix:**
- Wrapped homeView() content in bordered panel (like scout's file list panel)
- Added proper header "🏠 Dashboard" with consistent styling
- Removed nested box styling, simplified to clean sectioned layout
- Now matches the layout and structure of other tabs/panels

**Result:**
- View code structure now matches scout's clean, modular approach
- Better terminal size handling prevents layout issues
- Home page now properly aligned and consistent with other tabs
- Easier to maintain and extend with new view modes

### 2026-01-08 - Bug Fixes & Gamification Removal

**Bug Fixes:**
- Fixed unchecked json.Unmarshal error in storage.go that could cause silent data corruption
- Removed all gamification system (levels, points, achievements, daily streaks)
- Kept task streaks functionality (showing consecutive days completing tasks)

**Code Cleanup:**
- Simplified homeView() to show only essential stats and active task streaks
- Removed Achievement and Gamification types from model
- Cleaned up unused imports
- gamification.go now only contains updateTaskStreak() function

**Result:**
- UI is cleaner and less cluttered
- Focus on task tracking without game mechanics
- Task streaks still motivate consistency

### 2026-01-08 - Major Refactor & Distribution Setup (COMPLETED)

**Architecture Improvements:**
- Split monolithic main.go (2080 lines) into modular structure:
  - model.go - Data structures and state
  - update.go - Event handling and business logic
  - view.go - Rendering and UI
  - helpers.go - Utility functions and styles
  - storage.go - Data persistence
  - gamification.go - Achievement and streak logic
  - main.go - Clean entry point (16 lines)
- Much easier to navigate and maintain

**Distribution & Installation:**
- Created GitHub Actions workflow (.github/workflows/release.yml)
  - Automated multi-platform builds (Linux, macOS, Windows for amd64/arm64)
  - Generates checksums for security
  - Auto-creates GitHub releases on git tags
- Created install.sh script for one-liner installation
  - Auto-detects OS and architecture
  - Downloads pre-compiled binary
  - No Go toolchain required
- Updated README with multiple install options (curl script, binaries, go install)

**UX Enhancements:**
- Added help screen (press ?) with full keyboard shortcuts reference
- Polished dashboard with boxed sections:
  - Centered, styled stats header (Level, Points, Streak)
  - Progress box with aligned stats
  - Achievements box with unlock status
  - Active task streaks box
  - Warning boxes for rolling todos and expired reminders
  - Active reminders box with live countdowns
- Implemented modal-style confirmation dialogs
  - Centered overlay with rounded border
  - Clear visual hierarchy
  - Keyboard hints [y/n]

**Result:**
- App feels professional and distributable
- Clean architecture makes future changes easier
- Ready for GitHub releases and wider distribution
