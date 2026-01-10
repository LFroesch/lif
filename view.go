package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

func (m *model) dailyRows() []table.Row {
	rows := []table.Row{}
	sortDailies(m.data.Dailies, m.sortColumn[0], m.sortAscending[0])
	for _, daily := range m.data.Dailies {
		priority := daily.Priority
		if priority == "" {
			priority = "MEDIUM"
		}
		priority = strings.ToUpper(priority)

		var displayPriority string
		switch priority {
		case "HIGH":
			displayPriority = "HIGH"
		case "MEDIUM":
			displayPriority = "MEDIUM"
		case "LOW":
			displayPriority = "LOW"
		default:
			displayPriority = "MEDIUM"
		}

		// Status: only color the status column, not the whole row
		status := daily.Status
		if status == "DONE" {
			status = statusDoneStyle.Render("✓ DONE")
		} else {
			status = "INCOMPLETE"
		}

		streakDisplay := fmt.Sprintf("%d days", daily.CurrentStreak)
		if daily.CurrentStreak > 0 {
			streakDisplay = fmt.Sprintf("%d days 🔥", daily.CurrentStreak)
		}

		rows = append(rows, table.Row{
			normalizeText(daily.Task),
			displayPriority,
			normalizeText(daily.Category),
			streakDisplay,
			fmt.Sprintf("%d", daily.BestStreak),
			status,
		})
	}
	return rows
}

func (m *model) rollingRows() []table.Row {
	rows := []table.Row{}
	sortRollingTodos(m.data.RollingTodos, m.sortColumn[1], m.sortAscending[1])
	for _, todo := range m.data.RollingTodos {
		priority := todo.Priority
		if priority == "" {
			priority = "MEDIUM"
		}
		priority = strings.ToUpper(priority)

		var displayPriority string
		switch priority {
		case "HIGH":
			displayPriority = "HIGH"
		case "MEDIUM":
			displayPriority = "MEDIUM"
		case "LOW":
			displayPriority = "LOW"
		default:
			displayPriority = "MEDIUM"
		}

		rows = append(rows, table.Row{
			normalizeText(todo.Task),
			displayPriority,
			normalizeText(todo.Category),
			todo.Deadline,
		})
	}
	return rows
}

func (m *model) reminderRows() []table.Row {
	rows := []table.Row{}
	// Reminders aren't sortable, just display in order
	for _, reminder := range m.data.Reminders {
		// Display countdown/alarm time
		displayTime := reminder.AlarmOrCountdown
		if reminder.Status == "paused" && reminder.PausedRemaining > 0 {
			// Show paused remaining time
			if reminder.IsCountdown {
				displayTime = fmt.Sprintf("%s (PAUSED %s)", reminder.AlarmOrCountdown, reminder.PausedRemaining.Truncate(time.Second))
			} else {
				displayTime = fmt.Sprintf("%s (PAUSED)", reminder.AlarmOrCountdown)
			}
		} else if !reminder.TargetTime.IsZero() {
			remaining := time.Until(reminder.TargetTime)
			if remaining > 0 {
				if reminder.IsCountdown {
					displayTime = fmt.Sprintf("%s (%s)", reminder.AlarmOrCountdown, remaining.Truncate(time.Second))
				} else {
					displayTime = fmt.Sprintf("%s (%s)", reminder.AlarmOrCountdown, reminder.TargetTime.Format("15:04"))
				}
			} else {
				displayTime = fmt.Sprintf("%s (EXPIRED)", reminder.AlarmOrCountdown)
			}
		}

		rows = append(rows, table.Row{
			normalizeText(reminder.Reminder),
			normalizeText(reminder.Note),
			displayTime,
		})
	}
	return rows
}

func (m *model) filterReference() {
	query := strings.ToLower(m.searchInput.Value())
	if query == "" {
		m.filteredRef = m.data.Reference
		return
	}

	m.filteredRef = []ReferenceItem{}
	for _, item := range m.data.Reference {
		// Search in all fields
		if strings.Contains(strings.ToLower(item.Lang), query) ||
			strings.Contains(strings.ToLower(item.Command), query) ||
			strings.Contains(strings.ToLower(item.Usage), query) ||
			strings.Contains(strings.ToLower(item.Example), query) ||
			strings.Contains(strings.ToLower(item.Meaning), query) {
			m.filteredRef = append(m.filteredRef, item)
		}
	}
}

func (m *model) referenceRows() []table.Row {
	rows := []table.Row{}

	// Use filtered results if search is active
	itemsToShow := m.data.Reference
	if m.searchActive && m.searchInput.Value() != "" {
		itemsToShow = m.filteredRef
	}

	sortReference(itemsToShow, m.sortColumn[3], m.sortAscending[3])
	for _, item := range itemsToShow {
		rows = append(rows, table.Row{
			normalizeText(item.Lang),
			normalizeText(item.Command),
			normalizeText(item.Usage),
			normalizeText(item.Example),
			normalizeText(item.Meaning),
		})
	}
	return rows
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Show helpful message for very small terminals (like scout)
	if m.width < minTerminalWidth || m.height < minTerminalHeight {
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true).
			Padding(1)
		return warningStyle.Render(fmt.Sprintf(
			"Terminal too small: %dx%d\nMinimum: %dx%d\n\nPlease resize terminal or zoom out",
			m.width, m.height, minTerminalWidth, minTerminalHeight,
		))
	}

	var content string

	// Header
	header := m.renderHeader()

	// Main content area
	switch {
	case m.confirmDelete:
		content = m.renderConfirmDeleteView()
	case m.editing:
		content = m.editView()
	case m.showHelp:
		content = m.helpView()
	case m.activeTab == 1:
		content = m.homeView()
	case m.activeTab == 5:
		content = m.referenceView()
	default:
		// Table content for other tabs
		content = m.tables[m.activeTab-2].View()
	}

	// Status bar
	statusBar := m.renderStatusBar()

	// Special handling for modal dialogs (confirmDelete, editing, help)
	// These should be overlaid on the normal view
	if m.confirmDelete {
		return content
	}

	// Combine all sections
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		content,
		statusBar,
	)
}

func (m model) renderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(m.width)

	title := "📋 lif - lucas is forgetful"

	// Tab headers
	tabs := []string{}
	tabNames := []string{"[1] Home", "[2] Dailies", "[3] Rolling", "[4] Reminders", "[5] Reference"}

	for i, name := range tabNames {
		if i+1 == m.activeTab {
			tabs = append(tabs, activeTabStyle.Render(name))
		} else {
			tabs = append(tabs, tabStyle.Render(name))
		}
	}

	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(title),
		"",
		tabRow,
		"",
	)
}

func (m model) renderStatusBar() string {
	// Status bar with commands
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Width(m.width)

	var commands []string
	if m.activeTab == 1 {
		commands = append(commands, keyStyle.Render("1-5")+colonStyle.Render(": ")+actionStyle.Render("navigate"))
	} else {
		commands = append(commands, keyStyle.Render("↑↓")+colonStyle.Render(": ")+actionStyle.Render("navigate"))
		commands = append(commands, keyStyle.Render("e")+colonStyle.Render(": ")+actionStyle.Render("edit"))
		commands = append(commands, keyStyle.Render("n/a")+colonStyle.Render(": ")+actionStyle.Render("add"))
		commands = append(commands, keyStyle.Render("d")+colonStyle.Render(": ")+actionStyle.Render("delete"))
		if m.activeTab == 2 {
			commands = append(commands, keyStyle.Render("space/enter")+colonStyle.Render(": ")+actionStyle.Render("toggle done"))
			commands = append(commands, keyStyle.Render("s")+colonStyle.Render(": ")+actionStyle.Render("sort"))
		}
		if m.activeTab == 3 {
			commands = append(commands, keyStyle.Render("s")+colonStyle.Render(": ")+actionStyle.Render("sort"))
		}
		if m.activeTab == 5 {
			commands = append(commands, keyStyle.Render("/")+colonStyle.Render(": ")+actionStyle.Render("search"))
			commands = append(commands, keyStyle.Render("s")+colonStyle.Render(": ")+actionStyle.Render("sort"))
		}
		if m.activeTab == 4 {
			commands = append(commands, keyStyle.Render("s")+colonStyle.Render(": ")+actionStyle.Render("start/resume"))
			commands = append(commands, keyStyle.Render("p")+colonStyle.Render(": ")+actionStyle.Render("pause"))
			commands = append(commands, keyStyle.Render("r")+colonStyle.Render(": ")+actionStyle.Render("reset"))
		}
	}
	commands = append(commands, keyStyle.Render("?")+colonStyle.Render(": ")+actionStyle.Render("help"))
	commands = append(commands, keyStyle.Render("q")+colonStyle.Render(": ")+actionStyle.Render("quit"))
	commandRow := strings.Join(commands, bulletStyle.Render(" • "))

	// Status message (no expiry)
	if m.statusMsg != "" {
		statusMsgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.statusColor))
		commandRow += "\n> " + statusMsgStyle.Render(m.statusMsg)
	}

	return statusStyle.Render(commandRow)
}

func (m model) renderConfirmDeleteView() string {
	dialogWidth := 60
	dialogHeight := 10

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 2).
		Background(lipgloss.Color("235")).
		Width(dialogWidth).
		Height(dialogHeight)

	modalTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")).
		Render("⚠️  Confirm Deletion")

	modalContent := fmt.Sprintf("\n%s\n\nAre you sure you want to delete:\n\n  %s\n\n%s  %s\n",
		modalTitle,
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226")).Render(m.deleteTarget),
		keyStyle.Render("[y]")+" "+actionStyle.Render("Confirm"),
		keyStyle.Render("[n]")+" "+actionStyle.Render("Cancel"))

	modal := modalStyle.Render(modalContent)

	// Center the dialog
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		modal,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	)
}

func (m model) homeView() string {
	// Calculate available height for content
	availableHeight := m.height - uiOverhead
	if availableHeight < 3 {
		availableHeight = 3
	}

	// Content area
	contentStyle := lipgloss.NewStyle().
		Padding(0, 1)

	// Show summary stats
	totalDailies := len(m.data.Dailies)
	completedDailies := 0
	for _, daily := range m.data.Dailies {
		if daily.Status == "DONE" {
			completedDailies++
		}
	}

	var contentParts []string

	// Task Stats
	progressContent := statusDoneStyle.Render("📊 Your Progress") + "\n"
	progressContent += fmt.Sprintf("  Daily Tasks:         %d total, %d completed today\n", totalDailies, completedDailies)
	progressContent += fmt.Sprintf("  Rolling Todos:       %d items\n", len(m.data.RollingTodos))
	progressContent += fmt.Sprintf("  Active Reminders:    %d\n", len(m.data.Reminders))
	contentParts = append(contentParts, progressContent)

	// Rolling todos warning
	if len(m.data.RollingTodos) > 0 {
		todoWarning := "\n" + priorityHighStyle.Render("⚠️  Rolling Todos") + "\n"
		todoWarning += fmt.Sprintf("  You have %d rolling todos to complete\n", len(m.data.RollingTodos))
		contentParts = append(contentParts, todoWarning)
	}

	// Show expired reminders
	expiredReminders := []Reminder{}
	for _, reminder := range m.data.Reminders {
		if reminder.Status == "expired" {
			expiredReminders = append(expiredReminders, reminder)
		}
	}

	if len(expiredReminders) > 0 {
		expiredContent := "\n" + statusOverdueStyle.Render("⚠️ Expired Reminders") + "\n"
		for _, reminder := range expiredReminders {
			expiredContent += fmt.Sprintf("  • %s\n", reminder.Reminder)
		}
		contentParts = append(contentParts, expiredContent)
	}

	// Show active reminders with countdown
	activeReminders := []Reminder{}
	for _, reminder := range m.data.Reminders {
		if !reminder.TargetTime.IsZero() && (reminder.Status == "active" || reminder.Status == "paused") {
			activeReminders = append(activeReminders, reminder)
		}
	}

	// Sort by time remaining (soonest first)
	sort.Slice(activeReminders, func(i, j int) bool {
		iRemaining := time.Until(activeReminders[i].TargetTime)
		jRemaining := time.Until(activeReminders[j].TargetTime)

		// Handle paused reminders - use PausedRemaining for comparison
		if activeReminders[i].Status == "paused" && activeReminders[i].PausedRemaining > 0 {
			iRemaining = activeReminders[i].PausedRemaining
		}
		if activeReminders[j].Status == "paused" && activeReminders[j].PausedRemaining > 0 {
			jRemaining = activeReminders[j].PausedRemaining
		}

		// Sort by remaining time (ascending - soonest first)
		return iRemaining < jRemaining
	})

	if len(activeReminders) > 0 {
		reminderContent := "\n" + lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Render("🕐 Active Reminders") + "\n"

		for _, reminder := range activeReminders {
			statusIcon := "🕐"
			if reminder.Status == "paused" {
				statusIcon = "⏸️"
				// Show paused remaining time
				if reminder.PausedRemaining > 0 {
					if reminder.IsCountdown {
						reminderContent += fmt.Sprintf("  %s %s: %s (PAUSED)\n", statusIcon, reminder.Reminder, formatDuration(reminder.PausedRemaining))
					} else {
						reminderContent += fmt.Sprintf("  %s %s: PAUSED\n", statusIcon, reminder.Reminder)
					}
				} else {
					reminderContent += fmt.Sprintf("  %s %s: PAUSED\n", statusIcon, reminder.Reminder)
				}
			} else {
				// Active reminder - show live countdown
				remaining := time.Until(reminder.TargetTime)
				if remaining > 0 {
					if reminder.IsCountdown {
						reminderContent += fmt.Sprintf("  %s %s: %s\n", statusIcon, reminder.Reminder, formatDuration(remaining))
					} else {
						reminderContent += fmt.Sprintf("  %s %s: %s\n", statusIcon, reminder.Reminder, reminder.TargetTime.Format("15:04"))
					}
				} else {
					reminderContent += fmt.Sprintf("  ⚠️ %s: EXPIRED\n", reminder.Reminder)
				}
			}
		}
		contentParts = append(contentParts, reminderContent)
	}

	// Join all content parts
	content := strings.Join(contentParts, "")
	return contentStyle.Render(content)
}

func (m model) referenceView() string {
	// Reference tab with search
	var searchBox string
	if m.searchActive {
		searchBox = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Render("🔍 Search: ") + m.searchInput.View()
	} else {
		searchBox = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("Press / to search")
	}

	resultCount := ""
	if m.searchInput.Value() != "" {
		resultCount = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Render(fmt.Sprintf(" (%d results)", len(m.filteredRef)))
	}

	return searchBox + resultCount + "\n\n" + m.tables[3].View()
}

func (m model) helpView() string {
	availableHeight := m.height - uiOverhead
	if availableHeight < 3 {
		availableHeight = 3
	}

	// Reserve space for scroll indicators
	contentHeight := availableHeight - 2
	if contentHeight < 1 {
		contentHeight = 1
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("105")).
		Width(m.width - 4)

	header := headerStyle.Render("❓ Help")

	listStyle := lipgloss.NewStyle().
		Width(m.width-4).
		Padding(0, 1)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width - 2).
		Height(availableHeight + 2)

	// Build help content
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("105")).
		Bold(true)

	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)

	var allHelpContent []string

	// Global section
	allHelpContent = append(allHelpContent, sectionStyle.Render("Global:"))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s         Switch between tabs", keyStyle.Render("1-5")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s         Navigate tabs", keyStyle.Render("←/→")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s           Toggle this help screen", keyStyle.Render("?")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s   Quit application", keyStyle.Render("q / ctrl+c")))
	allHelpContent = append(allHelpContent, "")

	// Home section
	allHelpContent = append(allHelpContent, sectionStyle.Render("Home (Tab 1):"))
	allHelpContent = append(allHelpContent, "  View your stats and active reminders")
	allHelpContent = append(allHelpContent, "")

	// Daily Tasks section
	allHelpContent = append(allHelpContent, sectionStyle.Render("Daily Tasks (Tab 2):"))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s    Navigate list", keyStyle.Render("↑/↓ / j/k")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s  Toggle task completion (builds streaks!)", keyStyle.Render("space / enter")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s           Edit selected task", keyStyle.Render("e")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s         Add new task", keyStyle.Render("n / a")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s           Delete task", keyStyle.Render("d")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s           Cycle sort (Task/Priority/Category/Streak)", keyStyle.Render("s")))
	allHelpContent = append(allHelpContent, "")

	// Rolling Todos section
	allHelpContent = append(allHelpContent, sectionStyle.Render("Rolling Todos (Tab 3):"))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s    Navigate list", keyStyle.Render("↑/↓ / j/k")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s           Edit selected todo", keyStyle.Render("e")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s         Add new todo", keyStyle.Render("n / a")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s           Delete todo", keyStyle.Render("d")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s           Cycle sort", keyStyle.Render("s")))
	allHelpContent = append(allHelpContent, "")

	// Reminders section
	allHelpContent = append(allHelpContent, sectionStyle.Render("Reminders (Tab 4):"))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s    Navigate list", keyStyle.Render("↑/↓ / j/k")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s           Edit selected reminder", keyStyle.Render("e")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s         Add new reminder", keyStyle.Render("n / a")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s           Delete reminder", keyStyle.Render("d")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s           Start/resume reminder", keyStyle.Render("s")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s           Pause reminder", keyStyle.Render("p")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s           Reset reminder", keyStyle.Render("r")))
	allHelpContent = append(allHelpContent, "")

	// Reference section
	allHelpContent = append(allHelpContent, sectionStyle.Render("Reference (Tab 5):"))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s           Activate search", keyStyle.Render("/")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s         Clear search", keyStyle.Render("esc")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s           Cycle sort (Lang/Command)", keyStyle.Render("s")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s           Edit command", keyStyle.Render("e")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s         Add new command", keyStyle.Render("n / a")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s           Delete command", keyStyle.Render("d")))
	allHelpContent = append(allHelpContent, "")

	// Edit Mode section
	allHelpContent = append(allHelpContent, sectionStyle.Render("Edit Mode:"))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s         Next field", keyStyle.Render("tab")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s   Previous field", keyStyle.Render("shift+tab")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s       Save changes", keyStyle.Render("enter")))
	allHelpContent = append(allHelpContent, fmt.Sprintf("  %s         Cancel editing", keyStyle.Render("esc")))

	// Calculate visible range
	startIdx := m.helpScroll
	endIdx := m.helpScroll + contentHeight

	// Check if we need scroll indicators
	hasTopIndicator := startIdx > 0
	hasBottomIndicator := endIdx < len(allHelpContent)

	// Adjust for indicators
	if hasTopIndicator {
		contentHeight--
	}
	if hasBottomIndicator {
		contentHeight--
	}
	if contentHeight < 1 {
		contentHeight = 1
	}

	endIdx = startIdx + contentHeight
	if endIdx > len(allHelpContent) {
		endIdx = len(allHelpContent)
	}

	// Adjust scroll bounds
	if m.helpScroll > len(allHelpContent)-contentHeight {
		m.helpScroll = len(allHelpContent) - contentHeight
		if m.helpScroll < 0 {
			m.helpScroll = 0
		}
	}

	var displayLines []string
	if hasTopIndicator {
		displayLines = append(displayLines, "▲")
	}
	if startIdx < len(allHelpContent) {
		displayLines = append(displayLines, allHelpContent[startIdx:endIdx]...)
	}
	if hasBottomIndicator {
		displayLines = append(displayLines, "▼")
	}

	content := strings.Join(displayLines, "\n")
	listContent := listStyle.Render(content)

	combined := header + "\n" + listContent
	return borderStyle.Render(combined)
}

func (m model) editView() string {
	var fields []string
	var labels []string

	switch m.editingTab {
	case 2: // Dailies
		labels = []string{"Task:", "Priority:", "Category:", "Deadline:", "Status:"}
	case 3: // Rolling Todos
		labels = []string{"Task:", "Priority:", "Category:", "Deadline:"}
	case 4: // Reminders
		labels = []string{"Reminder:", "Note:", "Alarm/Countdown:"}
	case 5: // Reference
		labels = []string{"Lang:", "Command:", "Usage:", "Example:", "Meaning:"}
	}

	for i, input := range m.inputs {
		label := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render(labels[i])
		fields = append(fields, label+"\n"+input.View())
	}

	content := lipgloss.JoinVertical(lipgloss.Top, fields...)

	header := headerStyle.Render("✏️ Editing Mode")
	footer := keyStyle.Render("tab") + colonStyle.Render(": ") + actionStyle.Render("next field") + colonStyle.Render(" ") + bulletStyle.Render("•") + colonStyle.Render(" ") + keyStyle.Render("shift+tab") + colonStyle.Render(": ") + actionStyle.Render("prev field") + colonStyle.Render(" ") + bulletStyle.Render("•") + colonStyle.Render(" ") + keyStyle.Render("enter") + colonStyle.Render(": ") + actionStyle.Render("save") + colonStyle.Render(" ") + bulletStyle.Render("•") + colonStyle.Render(" ") + keyStyle.Render("esc") + colonStyle.Render(": ") + actionStyle.Render("cancel")

	return lipgloss.JoinVertical(lipgloss.Top,
		header,
		"",
		content,
		"",
		footer,
	)
}
