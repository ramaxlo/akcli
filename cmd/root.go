package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "ak",
	Short: "Automated Yocto build environment manager",
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "ak.toml", "path to config file")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
