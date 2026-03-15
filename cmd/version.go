package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version   = "v1.0.0"
	commit    = "unknown"
	buildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ak %s (commit: %s, built: %s)\n", version, commit, buildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
