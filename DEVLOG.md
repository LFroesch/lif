## DevLog
### 2026-03-23: Doc suite added
Added CLAUDE.md, agent_spec.md. Updated README to scout standard. Updated WORK.md with feature ideas.

### 2026-03-20: Fixed WSL2 notifications
sendNotification() now detects WSL and uses PowerShell toast notifications instead of notify-send.
Files: helpers.go

### 2026-01-09: Daily Task Streak System Rewrite
Fixed core streak bug: LastCompleted set before updateTaskStreak(). Simplified to use only LastCompleted with get3AMDay().

### 2026-01-09: Help page refactor + home page consistency
Scrollable help page matching scout style. Status bar background gap fix.

### 2026-01-08: Major refactor + distribution
Split 2080-line main.go into model/update/view/helpers/storage/gamification. GitHub Actions release workflow, install.sh.
