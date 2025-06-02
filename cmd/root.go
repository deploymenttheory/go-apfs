package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global output flags only
	verbose      bool
	quiet        bool
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:   "go-apfs",
	Short: "Cross-platform APFS filesystem explorer and extractor",
	Long: `go-apfs is a cross-platform, read-only command-line tool for exploring, 
extracting, recovering, and validating Apple File System (APFS) volumes.

Works directly with raw disks, partitions, or .dmg images without mounting
or relying on macOS. Ideal for data recovery, forensic analysis, and 
backup verification.

Commands:
  discover    Find files by name, extension, size, or content
  list        List volumes, snapshots, or files  
  extract     Extract files, directories, or volumes
  analyze     Analyze filesystem structure and integrity
  recover     Recover deleted files`,
	Version: "0.1.0-dev",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Only global output control flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress output except errors")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format (table, json, yaml)")
}

// GetVerbose returns the verbose flag value
func GetVerbose() bool {
	return verbose
}

// GetQuiet returns the quiet flag value
func GetQuiet() bool {
	return quiet
}

// GetOutputFormat returns the output format
func GetOutputFormat() string {
	return outputFormat
}
