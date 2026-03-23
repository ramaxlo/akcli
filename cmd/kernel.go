package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var kernelStartTime time.Time

var kernelCmd = &cobra.Command{
	Use:   "kernel",
	Short: "Kernel build commands",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		kernelStartTime = time.Now()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Elapsed time: %s\n", time.Since(kernelStartTime).Round(time.Second))
	},
}

func init() {
	rootCmd.AddCommand(kernelCmd)
}

func runCmd(name string, args ...string) error {
	fmt.Printf("Running: %s %s\n", name, strings.Join(args, " "))
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
