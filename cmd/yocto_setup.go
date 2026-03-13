package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ramaxlo/akcli/config"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Create yocto build directory by sourcing oe-init-build-env",
	RunE:  runSetup,
}

func init() {
	yoctoCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
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

	// Build the shell command: optionally export TEMPLATECONF, then source oe-init-build-env
	shellCmd := ""
	if cfg.Build.TemplateConf != "" {
		absTemplate, err := filepath.Abs(cfg.Build.TemplateConf)
		if err != nil {
			return fmt.Errorf("failed to resolve template_conf path: %w", err)
		}
		shellCmd = fmt.Sprintf("export TEMPLATECONF=%q && ", absTemplate)
	}
	shellCmd += fmt.Sprintf("source %q %s", absScript, buildDir)

	fmt.Printf("Sourcing oe-init-build-env with build dir '%s'...\n", buildDir)
	if cfg.Build.TemplateConf != "" {
		fmt.Printf("TEMPLATECONF=%s\n", cfg.Build.TemplateConf)
	}

	bash := exec.Command("bash", "-c", shellCmd)
	bash.Stdout = os.Stdout
	bash.Stderr = os.Stderr
	if err := bash.Run(); err != nil {
		return fmt.Errorf("failed to source oe-init-build-env: %w", err)
	}

	fmt.Println("Setup complete. Run 'ak yocto build' next.")
	return nil
}

func findPokyDir() (string, error) {
	// Common locations where poky might be checked out
	candidates := []string{
		"poky",
		"src/poky",
		"sources/poky",
		"layers/poky",
	}

	for _, c := range candidates {
		initScript := filepath.Join(c, "oe-init-build-env")
		if _, err := os.Stat(initScript); err == nil {
			return c, nil
		}
	}

	// Search for oe-init-build-env in current directory tree (one level deep)
	entries, err := os.ReadDir(".")
	if err != nil {
		return "", fmt.Errorf("failed to scan for poky directory: %w", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		initScript := filepath.Join(e.Name(), "oe-init-build-env")
		if _, err := os.Stat(initScript); err == nil {
			return e.Name(), nil
		}
	}

	return "", fmt.Errorf("could not find poky directory (oe-init-build-env not found)")
}
