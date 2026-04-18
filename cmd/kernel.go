package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var kernelStartTime time.Time
var kernelLogWriter *timestampWriter

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
	msg := fmt.Sprintf("Running: %s %s\n", name, strings.Join(args, " "))
	fmt.Print(msg)

	c := exec.Command(name, args...)
	if kernelLogWriter != nil {
		kernelLogWriter.Write([]byte(msg))
		c.Stdout = io.MultiWriter(os.Stdout, kernelLogWriter)
		c.Stderr = io.MultiWriter(os.Stderr, kernelLogWriter)
	} else {
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
	}
	return c.Run()
}
