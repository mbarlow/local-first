package monitoring

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type RequestLog struct {
	Timestamp time.Time `json:"timestamp"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	Status    int       `json:"status"`
	Duration  int64     `json:"duration_ms"`
	UserAgent string    `json:"user_agent,omitempty"`
	RemoteIP  string    `json:"remote_ip,omitempty"`
}

type Monitor struct {
	logFile string
	mu      sync.RWMutex
	logs    []RequestLog
}

func NewMonitor() *Monitor {
	logDir := filepath.Join(".", ".local-first")
	os.MkdirAll(logDir, 0755)
	
	return &Monitor{
		logFile: filepath.Join(logDir, "requests.jsonl"),
		logs:    make([]RequestLog, 0),
	}
}

func (m *Monitor) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a response writer wrapper to capture status code
		wrapper := &responseWrapper{
			ResponseWriter: w,
			statusCode:     200, // default
		}
		
		// Call the next handler
		next.ServeHTTP(wrapper, r)
		
		// Log the request
		duration := time.Since(start)
		
		reqLog := RequestLog{
			Timestamp: start,
			Method:    r.Method,
			Path:      r.URL.Path,
			Status:    wrapper.statusCode,
			Duration:  duration.Milliseconds(),
			UserAgent: r.UserAgent(),
			RemoteIP:  r.RemoteAddr,
		}
		
		m.logRequest(reqLog)
	})
}

func (m *Monitor) logRequest(reqLog RequestLog) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Add to in-memory logs (keep last 1000)
	m.logs = append(m.logs, reqLog)
	if len(m.logs) > 1000 {
		m.logs = m.logs[1:]
	}
	
	// Write to file
	go m.writeToFile(reqLog)
	
	// Print to console in development
	logMsg := fmt.Sprintf("%s %s %d %v",
		reqLog.Method,
		reqLog.Path,
		reqLog.Status,
		time.Duration(reqLog.Duration)*time.Millisecond,
	)
	fmt.Printf("[%s] %s\n", reqLog.Timestamp.Format("15:04:05"), logMsg)
}

func (m *Monitor) writeToFile(reqLog RequestLog) {
	file, err := os.OpenFile(m.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening log file: %v", err)
		return
	}
	defer file.Close()
	
	data, err := json.Marshal(reqLog)
	if err != nil {
		log.Printf("Error marshaling log: %v", err)
		return
	}
	
	file.Write(data)
	file.Write([]byte("\n"))
}

func (m *Monitor) GetRecentLogs(limit int) []RequestLog {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if limit <= 0 || limit > len(m.logs) {
		limit = len(m.logs)
	}
	
	if limit == 0 {
		return []RequestLog{}
	}
	
	start := len(m.logs) - limit
	result := make([]RequestLog, limit)
	copy(result, m.logs[start:])
	
	return result
}

func (m *Monitor) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if len(m.logs) == 0 {
		return map[string]interface{}{
			"total_requests": 0,
			"avg_duration":   0,
			"status_codes":   map[string]int{},
		}
	}
	
	statusCodes := make(map[string]int)
	var totalDuration int64
	
	for _, log := range m.logs {
		statusCodes[fmt.Sprintf("%d", log.Status)]++
		totalDuration += log.Duration
	}
	
	avgDuration := totalDuration / int64(len(m.logs))
	
	return map[string]interface{}{
		"total_requests": len(m.logs),
		"avg_duration":   avgDuration,
		"status_codes":   statusCodes,
	}
}

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}