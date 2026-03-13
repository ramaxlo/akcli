package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ramaxlo/akcli/config"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [extra bitbake flags...]",
	Short: "Source oe-init-build-env and run bitbake to build the image",
	RunE:  runBuild,
}

func init() {
	// Allow passing unknown flags through to bitbake
	buildCmd.DisableFlagParsing = true
	yoctoCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadCache()
	if err != nil {
		return err
	}

	pokyDir, err := findPokyDir()
	if err != nil {
		return err
	}

	initScript := filepath.Join(pokyDir, "oe-init-build-env")
	absScript, err := filepath.Abs(initScript)
	if err != nil {
		return fmt.Errorf("failed to resolve init script path: %w", err)
	}

	buildDir := cfg.Build.BuildDir
	target := cfg.Build.Target

	// Build the bitbake command with optional extra flags
	bbArgs := []string{target}
	if len(args) > 0 {
		bbArgs = append(bbArgs, args...)
	}

	// Source oe-init-build-env then run bitbake in the same shell
	shellCmd := fmt.Sprintf("source %q %s && bitbake %s",
		absScript, buildDir, strings.Join(bbArgs, " "))

	fmt.Printf("Building target '%s'...\n", target)
	if len(args) > 0 {
		fmt.Printf("Extra bitbake flags: %s\n", strings.Join(args, " "))
	}

	bash := exec.Command("bash", "-c", shellCmd)
	bash.Stdout = os.Stdout
	bash.Stderr = os.Stderr
	if err := bash.Run(); err != nil {
		return fmt.Errorf("bitbake build failed: %w", err)
	}

	fmt.Println("Build complete.")
	return nil
}
