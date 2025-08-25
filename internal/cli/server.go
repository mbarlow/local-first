package cli

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type ServerStatusMsg struct {
	Status ServerStatus
	PID    int
	Error  error
}

type ServerProcess struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

var currentServer *ServerProcess

func (m DashboardModel) checkServerStatus() tea.Cmd {
	return func() tea.Msg {
		port := m.server.Port
		
		// Check if port is in use
		if isPortInUse(port) {
			// Try to get PID
			pid := getProcessByPort(port)
			return ServerStatusMsg{
				Status: ServerRunning,
				PID:    pid,
			}
		}
		
		return ServerStatusMsg{
			Status: ServerStopped,
			PID:    0,
		}
	}
}

func (m DashboardModel) startServer() tea.Cmd {
	return func() tea.Msg {
		port := m.server.Port
		
		// Check if already running
		if isPortInUse(port) {
			return ServerStatusMsg{
				Status: ServerRunning,
				PID:    getProcessByPort(port),
			}
		}
		
		// Start the server
		ctx, cancel := context.WithCancel(context.Background())
		
		logger := GetLogger()
		
		// Build WASM first
		logger.Log(LogSystem, "cli", "Building WASM...")
		buildCmd := exec.Command("make", "wasm")
		if err := buildCmd.Run(); err != nil {
			cancel()
			logger.Log(LogError, "cli", fmt.Sprintf("Failed to build WASM: %v", err))
			return ServerStatusMsg{
				Status: ServerStopped,
				Error:  fmt.Errorf("failed to build WASM: %w", err),
			}
		}
		logger.Log(LogSystem, "cli", "WASM build completed")
		
		// Build server first
		logger.Log(LogSystem, "cli", "Building server...")
		buildServerCmd := exec.Command("make", "server")
		if err := buildServerCmd.Run(); err != nil {
			cancel()
			logger.Log(LogError, "cli", fmt.Sprintf("Failed to build server: %v", err))
			return ServerStatusMsg{
				Status: ServerStopped,
				Error:  fmt.Errorf("failed to build server: %w", err),
			}
		}
		logger.Log(LogSystem, "cli", "Server build completed")
		
		logger.Log(LogSystem, "cli", fmt.Sprintf("Starting server on port %d", port))
		
		// Start server using the built binary
		cmd := exec.CommandContext(ctx, "./bin/server", 
			"-dev", "-port", strconv.Itoa(port))
		
		// Set working directory to current directory
		cmd.Dir = "."
		
		// Set up process group so we can kill child processes  
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		
		// Capture server output and pipe to logger
		stdoutReader, err := cmd.StdoutPipe()
		if err != nil {
			cancel()
			logger.Log(LogError, "cli", fmt.Sprintf("Failed to create stdout pipe: %v", err))
			return ServerStatusMsg{
				Status: ServerStopped,
				Error:  fmt.Errorf("failed to create stdout pipe: %w", err),
			}
		}
		
		stderrReader, err := cmd.StderrPipe()
		if err != nil {
			cancel()
			logger.Log(LogError, "cli", fmt.Sprintf("Failed to create stderr pipe: %v", err))
			return ServerStatusMsg{
				Status: ServerStopped,
				Error:  fmt.Errorf("failed to create stderr pipe: %w", err),
			}
		}
		
		if err := cmd.Start(); err != nil {
			cancel()
			logger.Log(LogError, "cli", fmt.Sprintf("Failed to start server: %v", err))
			return ServerStatusMsg{
				Status: ServerStopped,
				Error:  fmt.Errorf("failed to start server: %w", err),
			}
		}
		
		logger.Log(LogSystem, "cli", fmt.Sprintf("Server started with PID %d", cmd.Process.Pid))
		
		// Start goroutines to read server output
		go NewStreamReader("server", LogInfo).Read(stdoutReader)
		go NewStreamReader("server", LogError).Read(stderrReader)
		
		currentServer = &ServerProcess{
			cmd:    cmd,
			cancel: cancel,
		}
		
		// Wait a moment for server to start
		time.Sleep(500 * time.Millisecond)
		
		logger.Log(LogSystem, "cli", "Server startup completed")
		
		return ServerStatusMsg{
			Status: ServerRunning,
			PID:    cmd.Process.Pid,
		}
	}
}

func (m DashboardModel) stopServer() tea.Cmd {
	return func() tea.Msg {
		logger := GetLogger()
		
		if currentServer != nil {
			logger.Log(LogSystem, "cli", "Stopping server...")
			
			// Store references before setting to nil
			cmd := currentServer.cmd
			cancel := currentServer.cancel
			
			// Cancel context
			cancel()
			
			// Kill process group
			if cmd != nil && cmd.Process != nil {
				pid := cmd.Process.Pid
				logger.Log(LogSystem, "cli", fmt.Sprintf("Terminating server process %d", pid))
				
				pgid, err := syscall.Getpgid(pid)
				if err == nil {
					syscall.Kill(-pgid, syscall.SIGTERM)
				}
				
				// Wait for process to exit
				go func() {
					cmd.Wait()
					logger.Log(LogSystem, "cli", "Server process has exited")
				}()
			}
			
			currentServer = nil
		} else {
			logger.Log(LogSystem, "cli", "No active server process, checking port...")
			
			// Try to kill any process using the port
			port := m.server.Port
			pid := getProcessByPort(port)
			if pid > 0 {
				logger.Log(LogSystem, "cli", fmt.Sprintf("Found process %d on port %d, terminating", pid, port))
				if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
					logger.Log(LogWarning, "cli", "SIGTERM failed, using SIGKILL")
					syscall.Kill(pid, syscall.SIGKILL)
				}
			} else {
				logger.Log(LogSystem, "cli", "No process found on port")
			}
		}
		
		logger.Log(LogSystem, "cli", "Server stopped")
		
		return ServerStatusMsg{
			Status: ServerStopped,
			PID:    0,
		}
	}
}

func (m DashboardModel) restartServer() tea.Cmd {
	return func() tea.Msg {
		logger := GetLogger()
		logger.Log(LogSystem, "cli", "Restarting server...")
		
		// Stop first
		m.stopServer()()
		
		// Wait a moment for cleanup
		logger.Log(LogSystem, "cli", "Waiting for cleanup...")
		time.Sleep(2 * time.Second)
		
		// Start again
		logger.Log(LogSystem, "cli", "Starting server again...")
		return m.startServer()()
	}
}

func getProcessByPort(port int) int {
	cmd := exec.Command("lsof", "-ti", ":"+strconv.Itoa(port))
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	
	pidStr := strings.TrimSpace(string(output))
	if pidStr == "" {
		return 0
	}
	
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0
	}
	
	return pid
}

func killProcessByPort(port int) error {
	cmd := exec.Command("lsof", "-ti", ":"+strconv.Itoa(port))
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	
	pidStr := strings.TrimSpace(string(output))
	if pidStr == "" {
		return nil // No process found
	}
	
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return err
	}
	
	// Send SIGTERM first
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		// If that fails, try SIGKILL
		return syscall.Kill(pid, syscall.SIGKILL)
	}
	
	return nil
}