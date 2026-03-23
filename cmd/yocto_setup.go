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

	if cfg.Build.TemplateConf == "" {
		return fmt.Errorf("template_conf is required in config")
	}

	absTemplate, err := filepath.Abs(cfg.Build.TemplateConf)
	if err != nil {
		return fmt.Errorf("failed to resolve template_conf path: %w", err)
	}

	// Build the shell command: export TEMPLATECONF, then source oe-init-build-env
	shellCmd := fmt.Sprintf("export TEMPLATECONF=%q && source %q %s", absTemplate, absScript, buildDir)

	// Resolve DL_DIR and SSTATE_DIR (default to project root)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	dlDir := filepath.Join(cwd, "downloads")
	if cfg.Build.DlDir != "" {
		dlDir, err = filepath.Abs(cfg.Build.DlDir)
		if err != nil {
			return fmt.Errorf("failed to resolve dl_dir path: %w", err)
		}
	}

	sstateDir := filepath.Join(cwd, "sstate-cache")
	if cfg.Build.SstateDir != "" {
		sstateDir, err = filepath.Abs(cfg.Build.SstateDir)
		if err != nil {
			return fmt.Errorf("failed to resolve sstate_dir path: %w", err)
		}
	}

	if dryRun {
		fmt.Printf("[dryrun] Would run: bash -c '%s'\n", shellCmd)
		fmt.Printf("[dryrun] Would append to %s/conf/local.conf:\n", buildDir)
		fmt.Printf("[dryrun]   DL_DIR = %q\n", dlDir)
		fmt.Printf("[dryrun]   SSTATE_DIR = %q\n", sstateDir)
		return nil
	}

	fmt.Printf("Sourcing oe-init-build-env with build dir '%s'...\n", buildDir)
	fmt.Printf("TEMPLATECONF=%s\n", cfg.Build.TemplateConf)
	fmt.Printf("Running: bash -c '%s'\n", shellCmd)

	bash := exec.Command("bash", "-c", shellCmd)
	bash.Stdout = os.Stdout
	bash.Stderr = os.Stderr
	if err := bash.Run(); err != nil {
		return fmt.Errorf("failed to source oe-init-build-env: %w", err)
	}

	// Append DL_DIR and SSTATE_DIR to local.conf
	localConf := filepath.Join(buildDir, "conf", "local.conf")
	if err := appendLocalConf(localConf, dlDir, sstateDir); err != nil {
		return err
	}

	fmt.Println("Setup complete. Run 'ak yocto build' next.")
	return nil
}

func appendLocalConf(path, dlDir, sstateDir string) error {
	var b strings.Builder
	b.WriteString("\n# Added by akcli\n")
	fmt.Fprintf(&b, "DL_DIR = %q\n", dlDir)
	fmt.Fprintf(&b, "SSTATE_DIR = %q\n", sstateDir)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open local.conf: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(b.String()); err != nil {
		return fmt.Errorf("failed to append to local.conf: %w", err)
	}

	fmt.Printf("DL_DIR = %q\n", dlDir)
	fmt.Printf("SSTATE_DIR = %q\n", sstateDir)
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
