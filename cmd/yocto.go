package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var startTime time.Time

var yoctoCmd = &cobra.Command{
	Use:   "yocto",
	Short: "Yocto build environment commands",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		startTime = time.Now()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Elapsed time: %s\n", time.Since(startTime).Round(time.Second))
	},
}

func init() {
	rootCmd.AddCommand(yoctoCmd)
}
