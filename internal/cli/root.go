package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "goforge",
	Short: "Generate a production-ready Go API project, no framework attached",
	Long: `GoForge generates the infrastructure every Go backend needs before
real development can start — config loading, structured logging, graceful
shutdown, PostgreSQL + sqlc wiring, Docker, and CI — so you can skip the
first day of setup and start writing your actual API.

The generated project has zero dependency on GoForge itself. It's ordinary,
idiomatic Go you're free to edit, restructure, or rip apart from line one.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}
