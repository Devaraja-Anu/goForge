package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var moduleFlag string

var newCmd = &cobra.Command{
	Use:   "new <project-name>",
	Short: "Create a new goforge project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		if moduleFlag == "" {
			fmt.Fprintln(os.Stderr, "error: --module is required")
			os.Exit(1)
		}
		fmt.Printf("Generating %s (module %s)...\n", projectName, moduleFlag)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVar(&moduleFlag, "module", "", "Go module path for the new project (required)")
	newCmd.MarkFlagRequired("module")
}
