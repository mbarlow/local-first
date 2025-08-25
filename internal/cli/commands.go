package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var DashboardCmd = &cobra.Command{
	Use:     "dashboard",
	Aliases: []string{"dash", "ui"},
	Short:   "Start the interactive dashboard",
	Long:    "Launch the Bubble Tea TUI dashboard to manage your local-first application",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize viper config
		initConfig()
		
		// Create and start the dashboard
		m := NewDashboardModel()
		p := tea.NewProgram(m, tea.WithAltScreen())
		
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running dashboard: %v\n", err)
			os.Exit(1)
		}
	},
}

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the development server",
	Long:  "Start the Go server for development",
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")
		dev, _ := cmd.Flags().GetBool("dev")
		
		fmt.Printf("Starting server on port %s (dev mode: %t)\n", port, dev)
		
		args = []string{"run", "cmd/server/main.go", "-port", port}
		if dev {
			args = append(args, "-dev")
		}
		
		serverCmd := exec.Command("go", args...)
		serverCmd.Stdout = os.Stdout
		serverCmd.Stderr = os.Stderr
		
		if err := serverCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
			os.Exit(1)
		}
	},
}

var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the WASM and server",
	Long:  "Build the WebAssembly module and/or server binary",
	Run: func(cmd *cobra.Command, args []string) {
		wasm, _ := cmd.Flags().GetBool("wasm")
		server, _ := cmd.Flags().GetBool("server")
		
		if !wasm && !server {
			// Default to building both
			wasm = true
			server = true
		}
		
		if wasm {
			fmt.Println("Building WASM...")
			if err := runMakeTarget("wasm"); err != nil {
				fmt.Fprintf(os.Stderr, "Error building WASM: %v\n", err)
				os.Exit(1)
			}
		}
		
		if server {
			fmt.Println("Building server...")
			if err := runMakeTarget("server"); err != nil {
				fmt.Fprintf(os.Stderr, "Error building server: %v\n", err)
				os.Exit(1)
			}
		}
		
		fmt.Println("Build complete!")
	},
}

func init() {
	// Serve command flags
	ServeCmd.Flags().StringP("port", "p", "8080", "Port to run the server on")
	ServeCmd.Flags().BoolP("dev", "d", true, "Run in development mode")
	
	// Build command flags  
	BuildCmd.Flags().Bool("wasm", false, "Build only WASM")
	BuildCmd.Flags().Bool("server", false, "Build only server")
}

func initConfig() {
	viper.SetConfigName("local")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/local-first")
	
	// Set defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.dev", true)
	viper.SetDefault("dashboard.refresh_interval", 1000)
	
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found is OK, we'll use defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		}
	}
}

func runMakeTarget(target string) error {
	cmd := exec.Command("make", target)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func isPortInUse(port int) bool {
	cmd := exec.Command("lsof", "-i", ":"+strconv.Itoa(port))
	err := cmd.Run()
	return err == nil
}