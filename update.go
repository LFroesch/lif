package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statusMsg:
		m.statusMsg = msg.message
		m.statusColor = msg.color
		m.statusExpiry = time.Now().Add(3 * time.Second)
		return m, nil

	case tickMsg:
		m.lastTick = time.Time(msg)

		// Check for daily task reset (runs every tick but only resets when needed)
		if resetDailyTasks(&m.data) {
			m.tables[0].SetRows(m.dailyRows())
			m.statusMsg = "🌅 Daily tasks reset at 3AM"
			m.statusColor = "82"
			m.statusExpiry = time.Now().Add(5 * time.Second)
			saveData(m.data)
		}

		// Check for reminder notifications (only for active reminders)
		for i, reminder := range m.data.Reminders {
			if !reminder.TargetTime.IsZero() && !reminder.Notified && reminder.Status == "active" && time.Now().After(reminder.TargetTime) {
				m.data.Reminders[i].Notified = true
				m.data.Reminders[i].Status = "expired"
				sendNotification("Reminder", reminder.Reminder)
				m.statusMsg = fmt.Sprintf("🔔 Reminder: %s", reminder.Reminder)
				m.statusColor = "226"
				m.statusExpiry = time.Now().Add(5 * time.Second)
				saveData(m.data)
			}
		}
		m.tables[2].SetRows(m.reminderRows())
		return m, tickCmd()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.adjustLayout()
		return m, nil

	case tea.KeyMsg:
		if m.editing {
			return m.handleEditingKeys(msg)
		}

		// Handle help screen
		if m.showHelp {
			switch msg.String() {
			case "?", "esc", "q":
				m.showHelp = false
				m.helpScroll = 0 // Reset scroll when closing
				return m, nil
			case "up", "k":
				if m.helpScroll > 0 {
					m.helpScroll--
				}
				return m, nil
			case "down", "j":
				m.helpScroll++
				return m, nil
			}
			return m, nil
		}

		// Handle search mode for Reference tab
		if m.searchActive && m.activeTab == 5 {
			switch msg.String() {
			case "esc":
				m.searchActive = false
				m.searchInput.Blur()
				m.searchInput.SetValue("")
				m.tables[3].SetRows(m.referenceRows())
				return m, nil
			case "enter":
				m.searchActive = false
				m.searchInput.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.filterReference()
				m.tables[3].SetRows(m.referenceRows())
				return m, cmd
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "1":
			m.activeTab = 1
		case "2":
			m.activeTab = 2
		case "3":
			m.activeTab = 3
		case "4":
			m.activeTab = 4
		case "5":
			m.activeTab = 5
		case "left":
			if m.activeTab > 1 {
				m.activeTab--
			} else if m.activeTab == 1 {
				m.activeTab = 5
			}
		case "right":
			if m.activeTab < 5 {
				m.activeTab++
			} else if m.activeTab == 5 {
				m.activeTab = 1
			}
		case "up", "k":
			if m.activeTab > 1 && m.activeTab < 6 {
				m.tables[m.activeTab-2], _ = m.tables[m.activeTab-2].Update(msg)
			}
		case "down", "j":
			if m.activeTab > 1 && m.activeTab < 6 {
				m.tables[m.activeTab-2], _ = m.tables[m.activeTab-2].Update(msg)
			}
		case "e":
			if m.activeTab > 1 && m.activeTab < 6 {
				m.startEditing()
			}
		case "n":
			if m.confirmDelete {
				m.confirmDelete = false
				m.deleteTarget = ""
				m.statusMsg = "Delete cancelled"
				m.statusColor = "86"
				m.statusExpiry = time.Now().Add(2 * time.Second)
			} else if m.activeTab > 1 && m.activeTab < 6 {
				m.addNew()
			}
		case "a":
			if m.activeTab > 1 && m.activeTab < 6 {
				m.addNew()
			}
		case "d", "delete":
			if m.activeTab > 1 && m.activeTab < 6 && !m.confirmDelete {
				m.confirmDeleteSelected()
			}
		case "y":
			if m.confirmDelete {
				m.deleteSelected()
				m.confirmDelete = false
				m.deleteTarget = ""
			}
		case "s":
			if m.activeTab == 4 {
				m.toggleReminderStatus("start")
			} else if m.activeTab == 2 || m.activeTab == 3 || m.activeTab == 5 {
				// Cycle sort for Dailies (tab 2), Rolling (tab 3), Reference (tab 5)
				m.cycleSortColumn()
			}
		case "p":
			if m.activeTab == 4 {
				m.toggleReminderStatus("pause")
			}
		case "r":
			if m.activeTab == 4 {
				m.toggleReminderStatus("reset")
			}
		case "/":
			// Activate search for Reference tab
			if m.activeTab == 5 {
				m.searchActive = true
				m.searchInput.Focus()
				return m, nil
			}
		case " ", "enter":
			// Toggle completion for dailies
			if m.activeTab == 2 {
				m.toggleCompletion()
			}

		}
	}

	return m, nil
}

func (m model) handleEditingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editing = false
		m.inputs = nil
		return m, showStatus("❌ Edit cancelled", "196")
	case "enter":
		m.saveEdit()
		m.editing = false
		m.inputs = nil
		return m, showStatus("✅ Changes saved", "82")
	case "tab":
		if len(m.inputs) > 0 {
			m.editingField = (m.editingField + 1) % len(m.inputs)
			for i := range m.inputs {
				m.inputs[i].Blur()
			}
			m.inputs[m.editingField].Focus()
		}
	case "shift+tab":
		if len(m.inputs) > 0 {
			m.editingField = (m.editingField - 1 + len(m.inputs)) % len(m.inputs)
			for i := range m.inputs {
				m.inputs[i].Blur()
			}
			m.inputs[m.editingField].Focus()
		}
	default:
		if len(m.inputs) > 0 {
			var cmd tea.Cmd
			m.inputs[m.editingField], cmd = m.inputs[m.editingField].Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m *model) startEditing() {
	m.editing = true
	m.editingTab = m.activeTab
	m.editingRow = m.tables[m.activeTab-2].Cursor()
	m.editingField = 0

	switch m.editingTab {
	case 2: // Dailies
		if m.editingRow < len(m.data.Dailies) {
			daily := m.data.Dailies[m.editingRow]
			m.inputs = make([]textinput.Model, 4)
			m.inputs[0] = textinput.New()
			m.inputs[0].SetValue(daily.Task)
			m.inputs[0].Focus()
			m.inputs[1] = textinput.New()
			m.inputs[1].SetValue(daily.Priority)
			m.inputs[2] = textinput.New()
			m.inputs[2].SetValue(daily.Category)
			m.inputs[3] = textinput.New()
			m.inputs[3].SetValue(daily.Deadline)
		}
	case 3: // Rolling Todos
		if m.editingRow < len(m.data.RollingTodos) {
			todo := m.data.RollingTodos[m.editingRow]
			m.inputs = make([]textinput.Model, 4)
			m.inputs[0] = textinput.New()
			m.inputs[0].SetValue(todo.Task)
			m.inputs[0].Focus()
			m.inputs[1] = textinput.New()
			m.inputs[1].SetValue(todo.Priority)
			m.inputs[2] = textinput.New()
			m.inputs[2].SetValue(todo.Category)
			m.inputs[3] = textinput.New()
			m.inputs[3].SetValue(todo.Deadline)
		}
	case 4: // Reminders
		if m.editingRow < len(m.data.Reminders) {
			reminder := m.data.Reminders[m.editingRow]
			m.inputs = make([]textinput.Model, 3)
			m.inputs[0] = textinput.New()
			m.inputs[0].SetValue(reminder.Reminder)
			m.inputs[0].Focus()
			m.inputs[1] = textinput.New()
			m.inputs[1].SetValue(reminder.Note)
			m.inputs[2] = textinput.New()
			m.inputs[2].SetValue(reminder.AlarmOrCountdown)
		}
	case 5: // Reference
		if m.editingRow < len(m.data.Reference) {
			item := m.data.Reference[m.editingRow]
			m.inputs = make([]textinput.Model, 5)
			m.inputs[0] = textinput.New()
			m.inputs[0].SetValue(item.Lang)
			m.inputs[0].Focus()
			m.inputs[1] = textinput.New()
			m.inputs[1].SetValue(item.Command)
			m.inputs[2] = textinput.New()
			m.inputs[2].SetValue(item.Usage)
			m.inputs[3] = textinput.New()
			m.inputs[3].SetValue(item.Example)
			m.inputs[4] = textinput.New()
			m.inputs[4].SetValue(item.Meaning)
		}
	}
}

func (m *model) addNew() {
	m.editing = true
	m.editingTab = m.activeTab
	m.editingRow = -1 // Indicates new item
	m.editingField = 0

	switch m.activeTab {
	case 2: // Dailies
		m.inputs = make([]textinput.Model, 4)
		for i := range m.inputs {
			m.inputs[i] = textinput.New()
		}
		m.inputs[0].Focus()
	case 3: // Rolling Todos
		m.inputs = make([]textinput.Model, 4)
		for i := range m.inputs {
			m.inputs[i] = textinput.New()
		}
		m.inputs[0].Focus()
	case 4: // Reminders
		m.inputs = make([]textinput.Model, 3)
		for i := range m.inputs {
			m.inputs[i] = textinput.New()
		}
		m.inputs[0].Focus()
	case 5: // Reference
		m.inputs = make([]textinput.Model, 5)
		for i := range m.inputs {
			m.inputs[i] = textinput.New()
		}
		m.inputs[0].Focus()
	}
}

func (m *model) saveEdit() {
	switch m.editingTab {
	case 2: // Dailies
		if m.editingRow == -1 {
			// New item
			newDaily := Daily{
				ID:            len(m.data.Dailies) + 1,
				Task:          normalizeText(m.inputs[0].Value()),
				Priority:      normalizePriority(m.inputs[1].Value()),
				Category:      normalizeText(m.inputs[2].Value()),
				Deadline:      m.inputs[3].Value(),
				Status:        "INCOMPLETE",
				LastCompleted: time.Time{},
			}
			m.data.Dailies = append(m.data.Dailies, newDaily)
		} else {
			// Edit existing
			m.data.Dailies[m.editingRow].Task = normalizeText(m.inputs[0].Value())
			m.data.Dailies[m.editingRow].Priority = normalizePriority(m.inputs[1].Value())
			m.data.Dailies[m.editingRow].Category = normalizeText(m.inputs[2].Value())
			m.data.Dailies[m.editingRow].Deadline = m.inputs[3].Value()
		}
		m.tables[0].SetRows(m.dailyRows())
	case 3: // Rolling Todos
		if m.editingRow == -1 {
			newTodo := RollingTodo{
				ID:       len(m.data.RollingTodos) + 1,
				Task:     normalizeText(m.inputs[0].Value()),
				Priority: normalizePriority(m.inputs[1].Value()),
				Category: normalizeText(m.inputs[2].Value()),
				Deadline: m.inputs[3].Value(),
			}
			m.data.RollingTodos = append(m.data.RollingTodos, newTodo)
		} else {
			m.data.RollingTodos[m.editingRow].Task = normalizeText(m.inputs[0].Value())
			m.data.RollingTodos[m.editingRow].Priority = normalizePriority(m.inputs[1].Value())
			m.data.RollingTodos[m.editingRow].Category = normalizeText(m.inputs[2].Value())
			m.data.RollingTodos[m.editingRow].Deadline = m.inputs[3].Value()
		}
		m.tables[1].SetRows(m.rollingRows())
	case 4: // Reminders
		if m.editingRow == -1 {
			newReminder := Reminder{
				ID:               len(m.data.Reminders) + 1,
				Reminder:         normalizeText(m.inputs[0].Value()),
				Note:             normalizeText(m.inputs[1].Value()),
				AlarmOrCountdown: m.inputs[2].Value(),
				CreatedAt:        time.Now(),
				Notified:         false,
			}
			// Parse countdown or alarm
			if targetTime, isCountdown := parseCountdown(m.inputs[2].Value()); isCountdown {
				newReminder.TargetTime = targetTime
				newReminder.IsCountdown = true
				newReminder.Status = "active"
			} else if targetTime, isAlarm := parseAlarmTime(m.inputs[2].Value()); isAlarm {
				newReminder.TargetTime = targetTime
				newReminder.IsCountdown = false
				newReminder.Status = "active"
			}
			m.data.Reminders = append(m.data.Reminders, newReminder)
		} else {
			m.data.Reminders[m.editingRow].Reminder = normalizeText(m.inputs[0].Value())
			m.data.Reminders[m.editingRow].Note = normalizeText(m.inputs[1].Value())
			m.data.Reminders[m.editingRow].AlarmOrCountdown = m.inputs[2].Value()
			// Re-parse countdown or alarm when editing
			if targetTime, isCountdown := parseCountdown(m.inputs[2].Value()); isCountdown {
				m.data.Reminders[m.editingRow].TargetTime = targetTime
				m.data.Reminders[m.editingRow].IsCountdown = true
				m.data.Reminders[m.editingRow].Notified = false
				m.data.Reminders[m.editingRow].Status = "active"
			} else if targetTime, isAlarm := parseAlarmTime(m.inputs[2].Value()); isAlarm {
				m.data.Reminders[m.editingRow].TargetTime = targetTime
				m.data.Reminders[m.editingRow].IsCountdown = false
				m.data.Reminders[m.editingRow].Notified = false
				m.data.Reminders[m.editingRow].Status = "active"
			}
		}
		m.tables[2].SetRows(m.reminderRows())
	case 5: // Reference
		if m.editingRow == -1 {
			newItem := ReferenceItem{
				ID:      len(m.data.Reference) + 1,
				Lang:    normalizeText(m.inputs[0].Value()),
				Command: normalizeText(m.inputs[1].Value()),
				Usage:   normalizeText(m.inputs[2].Value()),
				Example: normalizeText(m.inputs[3].Value()),
				Meaning: normalizeText(m.inputs[4].Value()),
			}
			m.data.Reference = append(m.data.Reference, newItem)
		} else {
			m.data.Reference[m.editingRow].Lang = normalizeText(m.inputs[0].Value())
			m.data.Reference[m.editingRow].Command = normalizeText(m.inputs[1].Value())
			m.data.Reference[m.editingRow].Usage = normalizeText(m.inputs[2].Value())
			m.data.Reference[m.editingRow].Example = normalizeText(m.inputs[3].Value())
			m.data.Reference[m.editingRow].Meaning = normalizeText(m.inputs[4].Value())
		}
		m.tables[3].SetRows(m.referenceRows())
	}

	saveData(m.data)
}

func (m *model) confirmDeleteSelected() {
	cursor := m.tables[m.activeTab-2].Cursor()
	var itemName string

	switch m.activeTab {
	case 2: // Dailies
		if cursor < len(m.data.Dailies) {
			itemName = m.data.Dailies[cursor].Task
		}
	case 3: // Rolling Todos
		if cursor < len(m.data.RollingTodos) {
			itemName = m.data.RollingTodos[cursor].Task
		}
	case 4: // Reminders
		if cursor < len(m.data.Reminders) {
			itemName = m.data.Reminders[cursor].Reminder
		}
	case 5: // Reference
		if cursor < len(m.data.Reference) {
			itemName = m.data.Reference[cursor].Command
		}
	}

	if itemName != "" {
		m.confirmDelete = true
		m.deleteTarget = itemName
	}
}

func (m *model) deleteSelected() {
	cursor := m.tables[m.activeTab-2].Cursor()

	switch m.activeTab {
	case 2: // Dailies
		if cursor < len(m.data.Dailies) {
			taskName := m.data.Dailies[cursor].Task
			m.data.Dailies = append(m.data.Dailies[:cursor], m.data.Dailies[cursor+1:]...)
			m.tables[0].SetRows(m.dailyRows())
			m.statusMsg = fmt.Sprintf("🗑️ Deleted: %s", taskName)
			m.statusColor = "196"
			m.statusExpiry = time.Now().Add(3 * time.Second)
		}
	case 3: // Rolling Todos
		if cursor < len(m.data.RollingTodos) {
			taskName := m.data.RollingTodos[cursor].Task
			m.data.RollingTodos = append(m.data.RollingTodos[:cursor], m.data.RollingTodos[cursor+1:]...)
			m.tables[1].SetRows(m.rollingRows())
			m.statusMsg = fmt.Sprintf("🗑️ Deleted: %s", taskName)
			m.statusColor = "196"
			m.statusExpiry = time.Now().Add(3 * time.Second)
		}
	case 4: // Reminders
		if cursor < len(m.data.Reminders) {
			reminderName := m.data.Reminders[cursor].Reminder
			m.data.Reminders = append(m.data.Reminders[:cursor], m.data.Reminders[cursor+1:]...)
			m.tables[2].SetRows(m.reminderRows())
			m.statusMsg = fmt.Sprintf("🗑️ Deleted: %s", reminderName)
			m.statusColor = "196"
			m.statusExpiry = time.Now().Add(3 * time.Second)
		}
	case 5: // Reference
		if cursor < len(m.data.Reference) {
			itemName := m.data.Reference[cursor].Command
			m.data.Reference = append(m.data.Reference[:cursor], m.data.Reference[cursor+1:]...)
			m.tables[3].SetRows(m.referenceRows())
			m.statusMsg = fmt.Sprintf("🗑️ Deleted: %s", itemName)
			m.statusColor = "196"
			m.statusExpiry = time.Now().Add(3 * time.Second)
		}
	}

	saveData(m.data)
}

func (m *model) cycleSortColumn() {
	var tableIdx int
	var maxColumns int

	switch m.activeTab {
	case 2: // Dailies
		tableIdx = 0
		maxColumns = 4 // Task, Priority, Category, Streak
	case 3: // Rolling
		tableIdx = 1
		maxColumns = 4 // Task, Priority, Category, Deadline
	case 5: // Reference
		tableIdx = 3
		maxColumns = 2 // Lang, Command
	default:
		return
	}

	// Cycle to next column
	m.sortColumn[tableIdx]++
	if m.sortColumn[tableIdx] >= maxColumns {
		m.sortColumn[tableIdx] = 0
		// Toggle direction when cycling back to first column
		m.sortAscending[tableIdx] = !m.sortAscending[tableIdx]
	}

	// Rebuild the affected table completely to ensure sort takes effect
	m.setupTables()

	// Show status message
	sortNames := map[int]map[int]string{
		0: {0: "Task", 1: "Priority", 2: "Category", 3: "Streak"},
		1: {0: "Task", 1: "Priority", 2: "Category", 3: "Deadline"},
		3: {0: "Lang", 1: "Command"},
	}

	direction := "↑"
	if !m.sortAscending[tableIdx] {
		direction = "↓"
	}

	m.statusMsg = fmt.Sprintf("Sorted by: %s %s", sortNames[tableIdx][m.sortColumn[tableIdx]], direction)
	m.statusColor = "86"
}

func (m *model) toggleReminderStatus(action string) {
	if m.activeTab != 4 || len(m.data.Reminders) == 0 {
		return
	}

	cursor := m.tables[2].Cursor()
	if cursor >= len(m.data.Reminders) {
		return
	}

	reminder := &m.data.Reminders[cursor]
	var statusMsg string
	var statusColor string

	switch action {
	case "start":
		if reminder.Status == "paused" {
			// Resume from paused state
			if reminder.PausedRemaining > 0 {
				reminder.TargetTime = time.Now().Add(reminder.PausedRemaining)
				reminder.PausedRemaining = 0
			}
			reminder.Status = "active"
			reminder.Notified = false
			statusMsg = fmt.Sprintf("▶️ Resumed: %s", reminder.Reminder)
			statusColor = "82"
		} else if reminder.Status == "inactive" {
			reminder.Status = "active"
			reminder.Notified = false
			// Re-parse the alarm/countdown
			if targetTime, isCountdown := parseCountdown(reminder.AlarmOrCountdown); isCountdown {
				reminder.TargetTime = targetTime
				reminder.IsCountdown = true
			} else if targetTime, isAlarm := parseAlarmTime(reminder.AlarmOrCountdown); isAlarm {
				reminder.TargetTime = targetTime
				reminder.IsCountdown = false
			}
			statusMsg = fmt.Sprintf("▶️ Started: %s", reminder.Reminder)
			statusColor = "82"
		} else {
			statusMsg = fmt.Sprintf("⚠️ %s is already active", reminder.Reminder)
			statusColor = "226"
		}

	case "pause":
		if reminder.Status == "active" {
			// Store remaining time when pausing
			if !reminder.TargetTime.IsZero() {
				reminder.PausedRemaining = time.Until(reminder.TargetTime)
				if reminder.PausedRemaining < 0 {
					reminder.PausedRemaining = 0
				}
			}
			reminder.Status = "paused"
			statusMsg = fmt.Sprintf("⏸️ Paused: %s", reminder.Reminder)
			statusColor = "226"
		} else {
			statusMsg = fmt.Sprintf("⚠️ %s is not active", reminder.Reminder)
			statusColor = "226"
		}

	case "reset":
		reminder.Status = "active"
		reminder.Notified = false
		reminder.PausedRemaining = 0 // Clear any paused time
		// Re-parse and reset the target time
		if targetTime, isCountdown := parseCountdown(reminder.AlarmOrCountdown); isCountdown {
			reminder.TargetTime = targetTime
			reminder.IsCountdown = true
		} else if targetTime, isAlarm := parseAlarmTime(reminder.AlarmOrCountdown); isAlarm {
			reminder.TargetTime = targetTime
			reminder.IsCountdown = false
		}
		statusMsg = fmt.Sprintf("🔄 Reset: %s", reminder.Reminder)
		statusColor = "82"
	}

	m.tables[2].SetRows(m.reminderRows())
	saveData(m.data)
	m.statusMsg = statusMsg
	m.statusColor = statusColor
	m.statusExpiry = time.Now().Add(3 * time.Second)
}

func (m *model) toggleCompletion() {
	if m.activeTab != 2 || len(m.data.Dailies) == 0 {
		return
	}

	cursor := m.tables[0].Cursor()
	if cursor >= len(m.data.Dailies) {
		return
	}

	current := m.data.Dailies[cursor].Status
	var newStatus string

	daily := &m.data.Dailies[cursor]

	switch current {
	case "DONE":
		newStatus = "INCOMPLETE"
		daily.LastCompleted = time.Time{} // Clear completion time
		m.statusMsg = fmt.Sprintf("Task marked as %s", newStatus)
		m.statusColor = "196"
	default:
		newStatus = "DONE"

		// Update task streak (before setting LastCompleted so it can check previous value)
		updateTaskStreak(daily)
		daily.LastCompleted = time.Now() // Record completion time

		if daily.CurrentStreak > 1 {
			m.statusMsg = fmt.Sprintf("✅ Task marked as %s! %d day streak! 🔥", newStatus, daily.CurrentStreak)
		} else {
			m.statusMsg = fmt.Sprintf("✅ Task marked as %s!", newStatus)
		}
		m.statusColor = "82"
	}

	daily.Status = newStatus
	m.tables[0].SetRows(m.dailyRows())
	saveData(m.data)
	m.statusExpiry = time.Now().Add(3 * time.Second)
}
