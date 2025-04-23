package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "cadence-codegen",
	Short: "Generate code from Cadence files",
	Long: `A tool for analyzing Cadence smart contracts and generating code.
It extracts transaction and script information, including parameters,
imports, and return types.`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// No need to register commands here as they register themselves in their own files
	rootCmd.SetVersionTemplate(`Version: {{.Version}}
Commit: ` + commit + `
Date: ` + date + `
`)
}
