package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

type ServerStatus int

const (
	ServerStopped ServerStatus = iota
	ServerStarting
	ServerRunning
	ServerStopping
)

func (s ServerStatus) String() string {
	switch s {
	case ServerStopped:
		return "Stopped"
	case ServerStarting:
		return "Starting..."
	case ServerRunning:
		return "Running"
	case ServerStopping:
		return "Stopping..."
	default:
		return "Unknown"
	}
}

type ServerInfo struct {
	Status ServerStatus
	Port   int
	PID    int
	Uptime time.Duration
}

type RequestLog struct {
	Timestamp time.Time
	Method    string
	Path      string
	Status    int
	Duration  time.Duration
}

type DashboardModel struct {
	server        ServerInfo
	requests      []RequestLog
	logs          []LogEntry
	selectedTab   int
	tabs          []string
	width, height int
	startTime     time.Time
	keyMap        KeyMap
	lastError     string
	showError     bool
}

type KeyMap struct {
	Start    key.Binding
	Stop     key.Binding
	Restart  key.Binding
	Refresh  key.Binding
	NextTab  key.Binding
	PrevTab  key.Binding
	Clear    key.Binding
	Quit     key.Binding
}

var DefaultKeyMap = KeyMap{
	Start: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "start server"),
	),
	Stop: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "stop server"),
	),
	Restart: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "restart server"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("f5", "ctrl+r"),
		key.WithHelp("f5", "refresh"),
	),
	NextTab: key.NewBinding(
		key.WithKeys("tab", "right"),
		key.WithHelp("tab", "next tab"),
	),
	PrevTab: key.NewBinding(
		key.WithKeys("shift+tab", "left"),
		key.WithHelp("shift+tab", "prev tab"),
	),
	Clear: key.NewBinding(
		key.WithKeys("c", "esc"),
		key.WithHelp("c", "clear error"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type tickMsg time.Time

func NewDashboardModel() DashboardModel {
	// Log CLI startup
	GetLogger().Log(LogSystem, "cli", "Dashboard started")
	
	return DashboardModel{
		server: ServerInfo{
			Status: ServerStopped,
			Port:   viper.GetInt("server.port"),
		},
		tabs:      []string{"Server", "Requests", "Logs"},
		startTime: time.Now(),
		keyMap:    DefaultKeyMap,
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		m.checkServerStatus(),
		m.tick(),
	)
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Start):
			if m.server.Status == ServerStopped {
				return m, m.startServer()
			}

		case key.Matches(msg, m.keyMap.Stop):
			if m.server.Status == ServerRunning {
				return m, m.stopServer()
			}

		case key.Matches(msg, m.keyMap.Restart):
			if m.server.Status == ServerRunning {
				return m, m.restartServer()
			}

		case key.Matches(msg, m.keyMap.NextTab):
			m.selectedTab = (m.selectedTab + 1) % len(m.tabs)

		case key.Matches(msg, m.keyMap.PrevTab):
			m.selectedTab = (m.selectedTab - 1 + len(m.tabs)) % len(m.tabs)

		case key.Matches(msg, m.keyMap.Refresh):
			return m, m.checkServerStatus()

		case key.Matches(msg, m.keyMap.Clear):
			m.showError = false
			m.lastError = ""
		}

	case tickMsg:
		m.updateUptime()
		return m, tea.Batch(
			m.checkServerStatus(),
			m.loadRequestLogs(),
			m.loadSystemLogs(),
			m.tick(),
		)

	case ServerStatusMsg:
		m.server.Status = msg.Status
		m.server.PID = msg.PID
		if msg.Error != nil {
			m.lastError = msg.Error.Error()
			m.showError = true
		} else {
			m.showError = false
		}
		if msg.Status == ServerRunning && m.startTime.IsZero() {
			m.startTime = time.Now()
		}
		if msg.Status == ServerStopped {
			m.startTime = time.Time{}
			m.server.Uptime = 0
		}

	case RequestLogsMsg:
		m.requests = msg.Logs
		
	case LogsUpdatedMsg:
		m.logs = msg.Logs
	}

	return m, nil
}

func (m DashboardModel) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	var content strings.Builder

	// Header
	header := m.renderHeader()
	content.WriteString(header)
	content.WriteString("\n\n")

	// Tabs
	tabs := m.renderTabs()
	content.WriteString(tabs)
	content.WriteString("\n\n")

	// Error message if present
	if m.showError {
		content.WriteString(m.renderError())
		content.WriteString("\n\n")
	}

	// Tab content
	switch m.selectedTab {
	case 0:
		content.WriteString(m.renderServerTab())
	case 1:
		content.WriteString(m.renderRequestsTab())
	case 2:
		content.WriteString(m.renderLogsTab())
	}

	// Footer
	content.WriteString("\n\n")
	content.WriteString(m.renderFooter())

	return content.String()
}

func (m DashboardModel) renderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		MarginLeft(2)

	return titleStyle.Render("🚀 Local-First Dashboard")
}

func (m DashboardModel) renderTabs() string {
	var tabs []string
	activeTabStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("36")).
		Background(lipgloss.Color("57")).
		Padding(0, 2)

	inactiveTabStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(0, 2)

	for i, tab := range m.tabs {
		if i == m.selectedTab {
			tabs = append(tabs, activeTabStyle.Render(tab))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(tab))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m DashboardModel) renderServerTab() string {
	var content strings.Builder

	// Server status
	statusStyle := lipgloss.NewStyle().
		Bold(true).
		Width(20)

	var statusColor lipgloss.Color
	switch m.server.Status {
	case ServerRunning:
		statusColor = lipgloss.Color("42") // Green
	case ServerStarting, ServerStopping:
		statusColor = lipgloss.Color("226") // Yellow
	default:
		statusColor = lipgloss.Color("196") // Red
	}

	content.WriteString(statusStyle.Render("Status:"))
	content.WriteString(" ")
	content.WriteString(
		lipgloss.NewStyle().
			Foreground(statusColor).
			Bold(true).
			Render(m.server.Status.String()),
	)
	content.WriteString("\n")

	content.WriteString(statusStyle.Render("Port:"))
	content.WriteString(" ")
	content.WriteString(strconv.Itoa(m.server.Port))
	content.WriteString("\n")

	if m.server.Status == ServerRunning {
		content.WriteString(statusStyle.Render("PID:"))
		content.WriteString(" ")
		content.WriteString(strconv.Itoa(m.server.PID))
		content.WriteString("\n")

		content.WriteString(statusStyle.Render("Uptime:"))
		content.WriteString(" ")
		content.WriteString(m.server.Uptime.Truncate(time.Second).String())
		content.WriteString("\n")

		content.WriteString(statusStyle.Render("URL:"))
		content.WriteString(" ")
		content.WriteString(
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("33")).
				Underline(true).
				Render(fmt.Sprintf("http://localhost:%d", m.server.Port)),
		)
		content.WriteString("\n")
	}

	return content.String()
}

func (m DashboardModel) renderRequestsTab() string {
	if len(m.requests) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("No requests yet... Start the server and visit http://localhost:" + strconv.Itoa(m.server.Port))
	}

	var content strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("33")).
		Width(80)
	
	content.WriteString(headerStyle.Render("TIME     METHOD PATH                    STATUS DURATION"))
	content.WriteString("\n")
	content.WriteString(strings.Repeat("─", 80))
	content.WriteString("\n")
	
	// Show last 15 requests
	recentReqs := m.requests
	if len(recentReqs) > 15 {
		recentReqs = recentReqs[len(recentReqs)-15:]
	}
	
	for _, req := range recentReqs {
		timeStr := req.Timestamp.Format("15:04:05")
		
		// Color code by status
		var statusColor lipgloss.Color
		switch {
		case req.Status >= 200 && req.Status < 300:
			statusColor = lipgloss.Color("42") // Green
		case req.Status >= 300 && req.Status < 400:
			statusColor = lipgloss.Color("226") // Yellow
		case req.Status >= 400:
			statusColor = lipgloss.Color("196") // Red
		default:
			statusColor = lipgloss.Color("241") // Gray
		}
		
		// Truncate path if too long
		path := req.Path
		if len(path) > 24 {
			path = path[:21] + "..."
		}
		
		// Duration color based on speed
		var durationColor lipgloss.Color
		ms := req.Duration.Milliseconds()
		switch {
		case ms < 10:
			durationColor = lipgloss.Color("42") // Green - fast
		case ms < 100:
			durationColor = lipgloss.Color("226") // Yellow - medium
		default:
			durationColor = lipgloss.Color("196") // Red - slow
		}
		
		content.WriteString(fmt.Sprintf("%s %-6s %-24s %s %s\n",
			timeStr,
			req.Method,
			path,
			lipgloss.NewStyle().Foreground(statusColor).Render(fmt.Sprintf("%-3d", req.Status)),
			lipgloss.NewStyle().Foreground(durationColor).Render(fmt.Sprintf("%4dms", ms)),
		))
	}
	
	// Summary stats
	if len(m.requests) > 0 {
		content.WriteString("\n")
		content.WriteString(strings.Repeat("─", 80))
		content.WriteString("\n")
		
		total := len(m.requests)
		var totalMs int64
		statusCounts := make(map[int]int)
		
		for _, req := range m.requests {
			totalMs += req.Duration.Milliseconds()
			statusCounts[req.Status/100*100]++
		}
		
		avgMs := totalMs / int64(total)
		
		summary := fmt.Sprintf("Total: %d requests • Avg: %dms • 2xx: %d • 3xx: %d • 4xx: %d • 5xx: %d",
			total, avgMs,
			statusCounts[200],
			statusCounts[300], 
			statusCounts[400],
			statusCounts[500],
		)
		
		content.WriteString(
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Render(summary),
		)
	}
	
	return content.String()
}

func (m DashboardModel) renderLogsTab() string {
	if len(m.logs) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("No logs yet... Start the server to see logs")
	}

	var content strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("33")).
		Width(80)
	
	content.WriteString(headerStyle.Render("TIME     LEVEL  SOURCE   MESSAGE"))
	content.WriteString("\n")
	content.WriteString(strings.Repeat("─", 80))
	content.WriteString("\n")
	
	// Show last 20 logs
	recentLogs := m.logs
	if len(recentLogs) > 20 {
		recentLogs = recentLogs[len(recentLogs)-20:]
	}
	
	for _, log := range recentLogs {
		timeStr := log.Timestamp.Format("15:04:05")
		
		// Color code by level
		var levelColor lipgloss.Color
		switch log.Level {
		case LogSystem:
			levelColor = lipgloss.Color("33") // Blue
		case LogInfo:
			levelColor = lipgloss.Color("42") // Green
		case LogWarning:
			levelColor = lipgloss.Color("226") // Yellow
		case LogError:
			levelColor = lipgloss.Color("196") // Red
		case LogDebug:
			levelColor = lipgloss.Color("241") // Gray
		default:
			levelColor = lipgloss.Color("241") // Gray
		}
		
		// Truncate message if too long
		message := log.Message
		if len(message) > 45 {
			message = message[:42] + "..."
		}
		
		content.WriteString(fmt.Sprintf("%s %-6s %-8s %s\n",
			timeStr,
			lipgloss.NewStyle().Foreground(levelColor).Render(fmt.Sprintf("%-6s", log.Level.String())),
			log.Source,
			message,
		))
	}
	
	// Summary
	if len(m.logs) > 0 {
		content.WriteString("\n")
		content.WriteString(strings.Repeat("─", 80))
		content.WriteString("\n")
		
		// Count by level
		counts := make(map[LogLevel]int)
		for _, log := range m.logs {
			counts[log.Level]++
		}
		
		summary := fmt.Sprintf("Total: %d logs • System: %d • Info: %d • Warn: %d • Error: %d",
			len(m.logs),
			counts[LogSystem],
			counts[LogInfo], 
			counts[LogWarning],
			counts[LogError],
		)
		
		content.WriteString(
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Render(summary),
		)
	}
	
	return content.String()
}

func (m DashboardModel) renderError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Background(lipgloss.Color("52")).
		Bold(true).
		Padding(0, 1).
		MarginLeft(2)

	return errorStyle.Render("❌ Error: " + m.lastError)
}

func (m DashboardModel) renderFooter() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	help := []string{
		"s: start",
		"x: stop", 
		"r: restart",
		"c: clear error",
		"tab: switch tabs",
		"q: quit",
	}

	return helpStyle.Render(strings.Join(help, " • "))
}

func (m DashboardModel) updateUptime() {
	if m.server.Status == ServerRunning && !m.startTime.IsZero() {
		m.server.Uptime = time.Since(m.startTime)
	}
}

func (m DashboardModel) tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}