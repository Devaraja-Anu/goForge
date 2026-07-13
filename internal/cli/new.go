package cli

import (
	"fmt"
	"path/filepath"

	blueprintfs "github.com/devaraja-anu/goforge"
	"github.com/devaraja-anu/goforge/internal/generator"
	"github.com/spf13/cobra"
)

var moduleFlag string

var newCmd = &cobra.Command{
	Use:   "new <project-name>",
	Short: "Generate a new Go API project",
	Long: `Creates a new directory containing a complete, ready-to-run Go API
service: Chi router, structured logging, graceful shutdown, PostgreSQL via
pgx + sqlc, golang-migrate migrations, Docker, and GitHub Actions CI.

No interactive prompts — module path is passed explicitly via --module.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		outputDir, err := filepath.Abs(projectName)
		if err != nil {
			return fmt.Errorf("resolving output path: %w", err)
		}

		err = generator.Generate(blueprintfs.FS, generator.Options{
			ProjectName: projectName,
			ModulePath:  moduleFlag,
			OutputDir:   outputDir,
		})
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}

		fmt.Printf("Generated %s (module %s)\n", projectName, moduleFlag)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVar(&moduleFlag, "module", "", "Go module path for the new project (required)")
	newCmd.MarkFlagRequired("module")
}
