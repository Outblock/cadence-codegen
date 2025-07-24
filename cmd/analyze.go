package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/outblock/cadence-codegen/internal/analyzer"
	"github.com/spf13/cobra"
)

var (
	includeBase64 bool
	resolveNested bool
	network       string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [input] [output]",
	Short: "Analyze Cadence files and generate JSON report",
	Long: `Analyze Cadence files and generate a JSON report.
The input can be either a single .cdc file or a directory containing .cdc files.
The output will be a JSON file containing the analysis result. If output is not specified, it defaults to 'cadence.json'.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]
		outputPath := "cadence.json"
		if len(args) > 1 {
			outputPath = args[1]
		}

		// Create analyzer
		a := analyzer.New()
		a.SetIncludeBase64(includeBase64)

		// Analyze directory
		err := a.AnalyzeDirectory(inputPath)
		if err != nil {
			return fmt.Errorf("failed to analyze directory: %w", err)
		}

		// Resolve nested types if requested
		if resolveNested {
			if network == "" {
				network = "mainnet" // default to mainnet
			}
			if err := a.ResolveNestedTypes(network); err != nil {
				fmt.Printf("Warning: failed to resolve nested types: %v\n", err)
			}
		}

		// Create output directory if it doesn't exist
		err = os.MkdirAll(filepath.Dir(outputPath), 0755)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Get the report
		report := a.GetReport()

		// Marshal to JSON with indentation
		jsonData, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}

		// Write to file
		err = os.WriteFile(outputPath, jsonData, 0644)
		if err != nil {
			return fmt.Errorf("failed to write JSON file: %w", err)
		}

		return nil
	},
}

func init() {
	analyzeCmd.Flags().BoolVar(&includeBase64, "base64", true, "Include base64-encoded Cadence files in the output")
	analyzeCmd.Flags().BoolVar(&resolveNested, "resolve-nested", true, "Resolve nested types by fetching contracts from chain")
	analyzeCmd.Flags().StringVar(&network, "network", "mainnet", "Network to use for resolving nested types (mainnet/testnet)")
	rootCmd.AddCommand(analyzeCmd)
}
