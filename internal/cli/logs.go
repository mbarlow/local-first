package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type RequestLogsMsg struct {
	Logs []RequestLog
}

func (m DashboardModel) loadRequestLogs() tea.Cmd {
	return func() tea.Msg {
		logFile := filepath.Join(".", ".local-first", "requests.jsonl")
		
		// Check if file exists
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			return RequestLogsMsg{Logs: []RequestLog{}}
		}
		
		data, err := os.ReadFile(logFile)
		if err != nil {
			return RequestLogsMsg{Logs: []RequestLog{}}
		}
		
		lines := strings.Split(string(data), "\n")
		var logs []RequestLog
		
		// Parse the last 50 lines (most recent logs)
		start := len(lines) - 51 // Extra line for empty line at end
		if start < 0 {
			start = 0
		}
		
		for i := start; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
			}
			
			var log struct {
				Timestamp time.Time `json:"timestamp"`
				Method    string    `json:"method"`
				Path      string    `json:"path"`
				Status    int       `json:"status"`
				Duration  int64     `json:"duration_ms"`
			}
			
			if err := json.Unmarshal([]byte(line), &log); err != nil {
				continue
			}
			
			logs = append(logs, RequestLog{
				Timestamp: log.Timestamp,
				Method:    log.Method,
				Path:      log.Path,
				Status:    log.Status,
				Duration:  time.Duration(log.Duration) * time.Millisecond,
			})
		}
		
		return RequestLogsMsg{Logs: logs}
	}
}