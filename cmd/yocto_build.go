package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ramaxlo/akcli/config"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:          "build [extra bitbake flags...]",
	Short:        "Source oe-init-build-env and run bitbake to build the image",
	RunE:         runBuild,
	SilenceUsage: true,
}

func init() {
	buildCmd.DisableFlagParsing = true
	yoctoCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, args []string) error {
	// Since flag parsing is disabled, manually handle our own flags
	var targetOverride string
	var bbArgs []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dryrun":
			dryRun = true
		case "--target":
			if i+1 < len(args) {
				i++
				targetOverride = args[i]
			} else {
				return fmt.Errorf("--target requires a value")
			}
		case "--help", "-h":
			fmt.Printf("Usage:\n  %s\n\n", cmd.UseLine())
			fmt.Printf("Flags:\n")
			fmt.Printf("  --target string   override the default build target from config\n")
			fmt.Printf("  --dryrun          display commands without executing them\n")
			fmt.Printf("  -h, --help        help for build\n")
			fmt.Printf("\nExtra bitbake flags are passed through directly to bitbake.\n")
			return nil
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
	if targetOverride != "" {
		target = targetOverride
	}

	// Build the bitbake command with optional extra flags
	allArgs := []string{target}
	if len(args) > 0 {
		allArgs = append(allArgs, args...)
	}

	// Source oe-init-build-env then run bitbake with MACHINE set
	shellCmd := fmt.Sprintf("source %q %s && MACHINE=%s bitbake %s",
		absScript, buildDir, cfg.Build.Machine, strings.Join(allArgs, " "))

	if dryRun {
		fmt.Printf("[dryrun] Would run: bash -c '%s'\n", shellCmd)
		return nil
	}

	fmt.Printf("Building target '%s'...\n", target)
	if len(args) > 0 {
		fmt.Printf("Extra bitbake flags: %s\n", strings.Join(args, " "))
	}
	fmt.Printf("Running: bash -c '%s'\n", shellCmd)

	var stdout io.Writer = os.Stdout
	var stderr io.Writer = os.Stderr

	if cfg.Build.BuildLog != "" {
		logFile, err := os.Create(cfg.Build.BuildLog)
		if err != nil {
			return fmt.Errorf("failed to create build log %s: %w", cfg.Build.BuildLog, err)
		}
		defer logFile.Close()

		tsw := &timestampWriter{w: logFile}
		defer tsw.Flush()

		stdout = io.MultiWriter(os.Stdout, tsw)
		stderr = io.MultiWriter(os.Stderr, tsw)
		fmt.Printf("Logging to: %s\n", cfg.Build.BuildLog)
	}

	bash := exec.Command("bash", "-c", shellCmd)
	bash.Stdout = stdout
	bash.Stderr = stderr
	if err := bash.Run(); err != nil {
		return fmt.Errorf("bitbake build failed: %w", err)
	}

	fmt.Println("Build complete.")
	return nil
}
