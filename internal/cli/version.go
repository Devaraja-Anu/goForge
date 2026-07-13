package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the goforge version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("goforge version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

}
