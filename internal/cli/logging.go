package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type LogLevel int

const (
	LogInfo LogLevel = iota
	LogWarning
	LogError
	LogDebug
	LogSystem
)

func (l LogLevel) String() string {
	switch l {
	case LogInfo:
		return "INFO"
	case LogWarning:
		return "WARN"
	case LogError:
		return "ERROR"
	case LogDebug:
		return "DEBUG"
	case LogSystem:
		return "SYSTEM"
	default:
		return "UNKNOWN"
	}
}

type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Source    string // "server", "wasm", "cli", etc.
	Message   string
}

type Logger struct {
	entries []LogEntry
	mu      sync.RWMutex
	logFile string
}

var globalLogger *Logger

func init() {
	logDir := filepath.Join(".", ".local-first")
	os.MkdirAll(logDir, 0755)
	
	globalLogger = &Logger{
		entries: make([]LogEntry, 0),
		logFile: filepath.Join(logDir, "cli.log"),
	}
}

func GetLogger() *Logger {
	return globalLogger
}

func (l *Logger) Log(level LogLevel, source, message string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Source:    source,
		Message:   strings.TrimSpace(message),
	}
	
	l.mu.Lock()
	l.entries = append(l.entries, entry)
	// Keep only last 500 entries in memory
	if len(l.entries) > 500 {
		l.entries = l.entries[1:]
	}
	l.mu.Unlock()
	
	// Write to file in background
	go l.writeToFile(entry)
}

func (l *Logger) writeToFile(entry LogEntry) {
	file, err := os.OpenFile(l.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	
	line := fmt.Sprintf("[%s] %s [%s] %s\n",
		entry.Timestamp.Format("15:04:05"),
		entry.Level.String(),
		entry.Source,
		entry.Message,
	)
	
	file.WriteString(line)
}

func (l *Logger) GetRecentLogs(limit int) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	if limit <= 0 || limit > len(l.entries) {
		limit = len(l.entries)
	}
	
	if limit == 0 {
		return []LogEntry{}
	}
	
	start := len(l.entries) - limit
	result := make([]LogEntry, limit)
	copy(result, l.entries[start:])
	
	return result
}

func (l *Logger) GetLogsBySource(source string, limit int) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	var filtered []LogEntry
	for _, entry := range l.entries {
		if entry.Source == source {
			filtered = append(filtered, entry)
		}
	}
	
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[len(filtered)-limit:]
	}
	
	return filtered
}

// StreamReader captures output from a command and logs it
type StreamReader struct {
	source string
	level  LogLevel
	logger *Logger
}

func NewStreamReader(source string, level LogLevel) *StreamReader {
	return &StreamReader{
		source: source,
		level:  level,
		logger: GetLogger(),
	}
}

func (sr *StreamReader) Read(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			sr.logger.Log(sr.level, sr.source, line)
		}
	}
}

// LogsUpdatedMsg is sent when new logs are available
type LogsUpdatedMsg struct {
	Logs []LogEntry
}

func (m DashboardModel) loadSystemLogs() tea.Cmd {
	return func() tea.Msg {
		logs := GetLogger().GetRecentLogs(50)
		return LogsUpdatedMsg{Logs: logs}
	}
}