package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var cfgFile string
var dryRun bool
var startTime time.Time

var rootCmd = &cobra.Command{
	Use:   "ak",
	Short: "Automated Yocto build environment manager",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		startTime = time.Now()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Elapsed time: %s\n", time.Since(startTime).Round(time.Second))
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "ak.toml", "path to config file")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dryrun", false, "display commands without executing them")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
