package main

import (
	"fmt"
	"os"

	"github.com/mbarlow/local-first/internal/cli"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "local",
	Short: "Local-first development tools",
	Long: `A CLI for managing your local-first Go WASM application.
	
Start the interactive dashboard, manage servers, and monitor your application
all from a beautiful terminal interface powered by Bubble Tea.`,
}

func main() {
	// Add commands
	rootCmd.AddCommand(cli.DashboardCmd)
	rootCmd.AddCommand(cli.ServeCmd)
	rootCmd.AddCommand(cli.BuildCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}