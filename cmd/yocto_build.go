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
	buildCmd.DisableFlagParsing = true
	yoctoCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, args []string) error {
	// Since flag parsing is disabled, manually handle our own flags
	var bbArgs []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dryrun":
			dryRun = true
		case "--help", "-h":
			return cmd.Help()
		default:
			bbArgs = append(bbArgs, args[i])
		}
	}
	args = bbArgs

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
	allArgs := []string{target}
	if len(args) > 0 {
		allArgs = append(allArgs, args...)
	}

	// Source oe-init-build-env then run bitbake in the same shell
	shellCmd := fmt.Sprintf("source %q %s && bitbake %s",
		absScript, buildDir, strings.Join(allArgs, " "))

	if dryRun {
		fmt.Printf("[dryrun] Would run: bash -c '%s'\n", shellCmd)
		return nil
	}

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
