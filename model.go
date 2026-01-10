package main

import (
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Data structures
type Daily struct {
	ID            int       `json:"id"`
	Task          string    `json:"task"`
	Priority      string    `json:"priority"`
	Category      string    `json:"category"`
	Deadline      string    `json:"deadline"`
	Status        string    `json:"status"`
	LastCompleted time.Time `json:"last_completed"`
	CurrentStreak int       `json:"current_streak"`
	BestStreak    int       `json:"best_streak"`
}

type RollingTodo struct {
	ID       int    `json:"id"`
	Task     string `json:"task"`
	Priority string `json:"priority"`
	Category string `json:"category"`
	Deadline string `json:"deadline"`
}

type Reminder struct {
	ID               int           `json:"id"`
	Reminder         string        `json:"reminder"`
	Note             string        `json:"note"`
	AlarmOrCountdown string        `json:"alarm_or_countdown"`
	Status           string        `json:"status"`
	CreatedAt        time.Time     `json:"created_at"`
	TargetTime       time.Time     `json:"target_time"`
	IsCountdown      bool          `json:"is_countdown"`
	Notified         bool          `json:"notified"`
	PausedRemaining  time.Duration `json:"paused_remaining"`
}

type ReferenceItem struct {
	ID      int    `json:"id"`
	Lang    string `json:"lang"`
	Command string `json:"command"`
	Usage   string `json:"usage"`
	Example string `json:"example"`
	Meaning string `json:"meaning"`
}

type AppData struct {
	Dailies      []Daily         `json:"dailies"`
	RollingTodos []RollingTodo   `json:"rolling_todos"`
	Reminders    []Reminder      `json:"reminders"`
	Reference    []ReferenceItem `json:"reference"`
}

type statusMsg struct {
	message string
	color   string
}

type tickMsg time.Time

type notificationMsg struct {
	reminder Reminder
}

// Terminal dimension constants
const (
	minTerminalWidth  = 60 // Minimum usable width
	minTerminalHeight = 20 // Minimum usable height
	uiOverhead        = 8  // Header (3) + status (2) + borders (2) + padding (1)
)

// Model
type model struct {
	activeTab      int
	tables         [4]table.Model
	data           AppData
	editing        bool
	editingTab     int
	editingRow     int
	editingField   int
	inputs         []textinput.Model
	statusMsg      string
	statusColor    string
	statusExpiry   time.Time
	width          int
	height         int
	lastTick       time.Time
	confirmDelete  bool
	deleteTarget   string
	sortColumn     [4]int  // Sort column for each table (Dailies, Rolling, Reminders, Reference)
	sortAscending  [4]bool // Sort direction for each table
	searchInput    textinput.Model
	searchActive   bool
	filteredRef    []ReferenceItem // Filtered reference items based on search
	showHelp       bool            // Toggle help screen
	helpScroll     int             // Help screen scroll position
}

func initialModel() model {
	m := model{
		activeTab:     1,
		data:          loadData(),
		statusColor:   "86",
		lastTick:      time.Now(),
		sortColumn:    [4]int{1, 1, 0, 0},              // Default sort: Priority for Dailies/Rolling, default for others
		sortAscending: [4]bool{true, true, true, true}, // All ascending by default
		searchActive:  false,
		filteredRef:   []ReferenceItem{},
		showHelp:      false,
	}

	// Initialize search input
	m.searchInput = textinput.New()
	m.searchInput.Placeholder = "Search commands..."
	m.searchInput.CharLimit = 50

	// Check for daily task reset on startup
	if resetDailyTasks(&m.data) {
		saveData(m.data)
	}

	m.setupTables()
	return m
}

func (m *model) setupTables() {
	// Calculate dynamic table height (leave space for header, tabs, status)
	tableHeight := m.height - 10
	if tableHeight < 10 {
		tableHeight = 10
	}

	// Tab 2: Dailies
	m.tables[0] = table.New(
		table.WithColumns([]table.Column{
			{Title: "Task", Width: 35},
			{Title: "Priority", Width: 12},
			{Title: "Category", Width: 18},
			{Title: "Streak", Width: 15},
			{Title: "Best", Width: 10},
			{Title: "Status", Width: 18},
		}),
		table.WithRows(m.dailyRows()),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	// Tab 3: Rolling Todos
	m.tables[1] = table.New(
		table.WithColumns([]table.Column{
			{Title: "Task", Width: 50},
			{Title: "Priority", Width: 12},
			{Title: "Category", Width: 20},
			{Title: "Deadline", Width: 20},
		}),
		table.WithRows(m.rollingRows()),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	// Tab 4: Reminders
	m.tables[2] = table.New(
		table.WithColumns([]table.Column{
			{Title: "Reminder", Width: 35},
			{Title: "Note", Width: 40},
			{Title: "Alarm/Countdown", Width: 40},
		}),
		table.WithRows(m.reminderRows()),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	// Tab 5: Reference
	m.tables[3] = table.New(
		table.WithColumns([]table.Column{
			{Title: "Lang", Width: 10},
			{Title: "Command", Width: 22},
			{Title: "Usage", Width: 30},
			{Title: "Example", Width: 30},
			{Title: "Meaning", Width: 30},
		}),
		table.WithRows(m.referenceRows()),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	// Apply modern table styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("255")) // White headers
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	for i := range m.tables {
		m.tables[i].SetStyles(s)
	}
}

func (m *model) adjustLayout() {
	if m.width == 0 || m.height == 0 {
		return
	}

	tableHeight := m.height - uiOverhead
	if tableHeight < 10 {
		tableHeight = 10
	}

	// Adjust table heights
	for i := range m.tables {
		m.tables[i].SetHeight(tableHeight)
	}
}

// Helper methods for safe dimensions (like scout)
func (m *model) getSafeWidth() int {
	if m.width < minTerminalWidth {
		return minTerminalWidth
	}
	return m.width
}

func (m *model) getSafeHeight() int {
	if m.height < minTerminalHeight {
		return minTerminalHeight
	}
	return m.height
}

// getContentHeight returns available height for content (total - UI overhead)
func (m *model) getContentHeight() int {
	availableHeight := m.getSafeHeight() - uiOverhead
	if availableHeight < 3 {
		availableHeight = 3
	}
	return availableHeight
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}
