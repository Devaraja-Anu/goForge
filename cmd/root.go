/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var verbose bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goforge",
	Short: "An opinionated package that handles the init",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Goforge is an app that generates all the average boilerplate code 
in a golang backend project, allowing you to kickstart the project asap.`,
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
