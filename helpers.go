package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Enhanced styles with better color coding
var (
	// Tab styles
	tabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Background(lipgloss.Color("236")).
			PaddingLeft(1).
			PaddingRight(1)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			PaddingLeft(1).
			PaddingRight(1)

	// Priority color styles
	priorityHighStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true) // Red
	priorityMedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true) // Yellow
	priorityLowStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)  // Green

	// Status color styles
	statusDoneStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)  // Green
	statusPendingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true) // Yellow
	statusOverdueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true) // Red

	// Command styles
	keyStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Background(lipgloss.Color("236"))  // Blue on gray
	actionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Background(lipgloss.Color("236"))  // Green on gray
	bulletStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(lipgloss.Color("236")) // Gray on gray
	colonStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(lipgloss.Color("236")) // Gray on gray

	// Header style
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))
)

func showStatus(msg string, color string) tea.Cmd {
	return func() tea.Msg {
		return statusMsg{message: msg, color: color}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func normalizeText(text string) string {
	return strings.ToLower(strings.TrimSpace(text))
}

func normalizePriority(priority string) string {
	norm := strings.ToUpper(strings.TrimSpace(priority))
	switch norm {
	case "HIGH", "H":
		return "HIGH"
	case "MEDIUM", "MED", "M":
		return "MEDIUM"
	case "LOW", "L":
		return "LOW"
	default:
		// Handle legacy values
		lower := strings.ToLower(norm)
		if strings.Contains(lower, "high") {
			return "HIGH"
		} else if strings.Contains(lower, "low") {
			return "LOW"
		}
		return "MEDIUM"
	}
}

func parseCountdown(countdownStr string) (time.Time, bool) {
	// Days format (1d, 5d, 20d)
	if strings.HasSuffix(countdownStr, "d") {
		dayStr := strings.TrimSuffix(countdownStr, "d")
		if days, err := strconv.Atoi(dayStr); err == nil {
			return time.Now().Add(time.Duration(days) * 24 * time.Hour), true
		}
	}

	// Weeks format (1w, 2w)
	if strings.HasSuffix(countdownStr, "w") {
		weekStr := strings.TrimSuffix(countdownStr, "w")
		if weeks, err := strconv.Atoi(weekStr); err == nil {
			return time.Now().Add(time.Duration(weeks) * 7 * 24 * time.Hour), true
		}
	}

	// Minutes format (1m, 30m, min)
	if strings.HasSuffix(countdownStr, "m") || strings.HasSuffix(countdownStr, "min") {
		minStr := strings.TrimSuffix(strings.TrimSuffix(countdownStr, "min"), "m")
		if minutes, err := strconv.Atoi(minStr); err == nil {
			return time.Now().Add(time.Duration(minutes) * time.Minute), true
		}
	}

	// Hours format (1h, 2h, hr)
	if strings.HasSuffix(countdownStr, "h") || strings.HasSuffix(countdownStr, "hr") {
		hourStr := strings.TrimSuffix(strings.TrimSuffix(countdownStr, "hr"), "h")
		if hours, err := strconv.Atoi(hourStr); err == nil {
			return time.Now().Add(time.Duration(hours) * time.Hour), true
		}
	}

	// Seconds format (1s, 30s, sec)
	if strings.HasSuffix(countdownStr, "s") || strings.HasSuffix(countdownStr, "sec") {
		secStr := strings.TrimSuffix(strings.TrimSuffix(countdownStr, "sec"), "s")
		if seconds, err := strconv.Atoi(secStr); err == nil {
			return time.Now().Add(time.Duration(seconds) * time.Second), true
		}
	}

	return time.Time{}, false
}

func parseAlarmTime(alarmStr string) (time.Time, bool) {
	now := time.Now()

	// Try 12-hour format first (1:50PM, 1:50 PM, 1:50pm, etc.)
	formats12 := []string{"3:04PM", "3:04 PM", "3:04pm", "3:04 pm"}
	for _, format := range formats12 {
		if t, err := time.Parse(format, alarmStr); err == nil {
			alarmTime := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
			if alarmTime.Before(now) {
				alarmTime = alarmTime.Add(24 * time.Hour)
			}
			return alarmTime, true
		}
	}

	// Try 24-hour format (15:04)
	if t, err := time.Parse("15:04", alarmStr); err == nil {
		alarmTime := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
		if alarmTime.Before(now) {
			alarmTime = alarmTime.Add(24 * time.Hour)
		}
		return alarmTime, true
	}

	return time.Time{}, false
}

func formatDuration(d time.Duration) string {
	// If over 8 hours, round to nearest hour
	if d > 8*time.Hour {
		hours := d.Round(time.Hour)
		if hours >= 24*time.Hour {
			days := int(hours / (24 * time.Hour))
			remaining := hours % (24 * time.Hour)
			if remaining == 0 {
				if days == 1 {
					return "1 day"
				}
				return fmt.Sprintf("%d days", days)
			} else {
				hours := int(remaining / time.Hour)
				if days == 1 {
					return fmt.Sprintf("1 day %dh", hours)
				}
				return fmt.Sprintf("%dd %dh", days, hours)
			}
		} else {
			hours := int(d.Round(time.Hour) / time.Hour)
			if hours == 1 {
				return "1 hour"
			}
			return fmt.Sprintf("%d hours", hours)
		}
	}

	// For under 8 hours, show precise time
	return d.Truncate(time.Second).String()
}

func isWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	// Check if we're in WSL by looking for WSL-specific environment variables or files
	if os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSLENV") != "" {
		return true
	}
	// Check for WSL filesystem marker
	if _, err := os.Stat("/proc/version"); err == nil {
		if data, err := os.ReadFile("/proc/version"); err == nil {
			return strings.Contains(string(data), "microsoft") || strings.Contains(string(data), "WSL")
		}
	}
	return false
}

func playNotificationSound() {
	// Try both mp3 and wav files
	soundFiles := []string{"assets/notification.mp3", "assets/notification.wav"}
	var soundFile string
	for _, file := range soundFiles {
		if _, err := os.Stat(file); err == nil {
			soundFile = file
			break
		}
	}

	// If no sound file exists, play system beep
	if soundFile == "" {
		if isWSL() {
			go exec.Command("powershell.exe", "-Command", "[console]::beep(800,200)").Run()
		} else {
			go exec.Command("printf", "\a").Run()
		}
		return
	}

	if isWSL() {
		// In WSL, just use Linux audio players if available
		players := [][]string{
			{"mpv", "--no-video", "--really-quiet", "--audio-buffer=1.0", soundFile},
			{"vlc", "--intf", "dummy", "--play-and-exit", soundFile},
			{"mplayer", "-really-quiet", soundFile},
			{"ffplay", "-nodisp", "-autoexit", "-v", "quiet", soundFile},
		}
		for _, cmd := range players {
			if _, err := exec.LookPath(cmd[0]); err == nil {
				go exec.Command(cmd[0], cmd[1:]...).Run()
				return
			}
		}
		// If no players available, just beep
		go exec.Command("powershell.exe", "-Command", "[console]::beep(800,200)").Run()
		return
	}

	switch runtime.GOOS {
	case "linux":
		// Try different audio players (in order of preference)
		players := [][]string{
			{"mpv", "--no-video", "--really-quiet", "--audio-buffer=1.0", soundFile},
			{"vlc", "--intf", "dummy", "--play-and-exit", soundFile},
			{"mplayer", "-really-quiet", soundFile},
			{"ffplay", "-nodisp", "-autoexit", "-v", "quiet", soundFile},
		}
		for _, cmd := range players {
			if _, err := exec.LookPath(cmd[0]); err == nil {
				go exec.Command(cmd[0], cmd[1:]...).Run()
				return
			}
		}
	case "darwin":
		// Use afplay on macOS
		go exec.Command("afplay", soundFile).Run()
	case "windows":
		// Use PowerShell to play sound on Windows
		go exec.Command("powershell", "-Command", fmt.Sprintf(`(New-Object Media.SoundPlayer "%s").PlaySync()`, soundFile)).Run()
	}
}

func sendNotification(title, message string) {
	// Play notification sound
	playNotificationSound()

	// Send system notification
	switch runtime.GOOS {
	case "linux":
		exec.Command("notify-send", title, message).Run()
	case "darwin":
		exec.Command("osascript", "-e", fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)).Run()
	case "windows":
		exec.Command("powershell", "-Command", fmt.Sprintf(`[System.Reflection.Assembly]::LoadWithPartialName('System.Windows.Forms'); [System.Windows.Forms.MessageBox]::Show('%s', '%s')`, message, title)).Run()
	}
}

func getMostRecent3AM() time.Time {
	now := time.Now()
	today3AM := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())

	// If current time is before 3AM today, use yesterday's 3AM
	if now.Before(today3AM) {
		return today3AM.AddDate(0, 0, -1)
	}

	// If current time is after 3AM today, use today's 3AM
	return today3AM
}

func resetDailyTasks(data *AppData) bool {
	mostRecent3AM := getMostRecent3AM()
	now := time.Now()
	resetOccurred := false

	for i := range data.Dailies {
		daily := &data.Dailies[i]
		// Reset to INCOMPLETE if task was completed before the most recent 3AM
		if daily.Status == "DONE" && daily.LastCompleted.Before(mostRecent3AM) {
			daily.Status = "INCOMPLETE"
			daily.LastCompleted = time.Time{} // Reset completion time
			daily.CurrentStreak = 0           // Reset streak
			resetOccurred = true
		}

		// Check if streak should be broken (task not completed in previous 3AM-based day)
		if daily.CurrentStreak > 0 && !daily.LastCompleted.IsZero() {
			lastCompletedDay := get3AMDay(daily.LastCompleted)
			yesterdayTime := now.Add(-24 * time.Hour)
			yesterday := get3AMDay(yesterdayTime)

			// If last completion was not yesterday or today, break the streak
			if lastCompletedDay != yesterday && lastCompletedDay != get3AMDay(now) {
				daily.CurrentStreak = 0
				resetOccurred = true
			}
		}
	}

	return resetOccurred
}

func sortDailies(items []Daily, column int, ascending bool) {
	pri := map[string]int{"HIGH": 0, "MEDIUM": 1, "LOW": 2}
	sort.Slice(items, func(i, j int) bool {
		var less bool
		switch column {
		case 0: // Task
			less = items[i].Task < items[j].Task
		case 1: // Priority
			iPri := strings.ToUpper(items[i].Priority)
			jPri := strings.ToUpper(items[j].Priority)
			if iPri == "" {
				iPri = "MEDIUM"
			}
			if jPri == "" {
				jPri = "MEDIUM"
			}
			less = pri[iPri] < pri[jPri]
		case 2: // Category
			less = strings.ToLower(items[i].Category) < strings.ToLower(items[j].Category)
		case 3: // Streak
			less = items[i].CurrentStreak < items[j].CurrentStreak
		default:
			less = items[i].Task < items[j].Task
		}
		if !ascending {
			return !less
		}
		return less
	})
}

func sortRollingTodos(items []RollingTodo, column int, ascending bool) {
	pri := map[string]int{"HIGH": 0, "MEDIUM": 1, "LOW": 2}
	sort.Slice(items, func(i, j int) bool {
		var less bool
		switch column {
		case 0: // Task
			less = items[i].Task < items[j].Task
		case 1: // Priority
			iPri := strings.ToUpper(items[i].Priority)
			jPri := strings.ToUpper(items[j].Priority)
			if iPri == "" {
				iPri = "MEDIUM"
			}
			if jPri == "" {
				jPri = "MEDIUM"
			}
			less = pri[iPri] < pri[jPri]
		case 2: // Category
			less = strings.ToLower(items[i].Category) < strings.ToLower(items[j].Category)
		case 3: // Deadline
			less = items[i].Deadline < items[j].Deadline
		default:
			less = items[i].Task < items[j].Task
		}
		if !ascending {
			return !less
		}
		return less
	})
}

func sortReference(items []ReferenceItem, column int, ascending bool) {
	sort.Slice(items, func(i, j int) bool {
		var less bool
		switch column {
		case 0: // Lang
			less = items[i].Lang < items[j].Lang
		case 1: // Command
			less = items[i].Command < items[j].Command
		default:
			less = items[i].Lang < items[j].Lang
		}
		if !ascending {
			return !less
		}
		return less
	})
}
