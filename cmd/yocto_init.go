package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/ramaxlo/akcli/config"
	"github.com/spf13/cobra"
)

const repoURL = "https://storage.googleapis.com/git-repo-downloads/repo"

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Fetch yocto manifest and set up repo tool",
	RunE:  runInit,
}

func init() {
	yoctoCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	fmt.Printf("Manifest URL: %s\n", cfg.Manifest.URL)
	if cfg.Manifest.Tag != "" {
		fmt.Printf("Tag:          %s\n", cfg.Manifest.Tag)
	} else {
		fmt.Printf("Branch:       %s\n", cfg.Manifest.Branch)
	}
	fmt.Printf("Distro:       %s\n", cfg.Build.Distro)
	fmt.Printf("Machine:      %s\n", cfg.Build.Machine)
	fmt.Printf("Target:       %s\n", cfg.Build.Target)

	revision := cfg.Manifest.Branch
	if cfg.Manifest.Tag != "" {
		revision = "refs/tags/" + cfg.Manifest.Tag
	}

	if dryRun {
		fmt.Println("[dryrun] Would ensure repo tool is available")
		fmt.Printf("[dryrun] Would run: repo init -u %s -b %s\n", cfg.Manifest.URL, revision)
		fmt.Println("[dryrun] Would cache config to .ak/config.cache.toml")
		return nil
	}

	repoPath, err := ensureRepoTool()
	if err != nil {
		return fmt.Errorf("failed to set up repo tool: %w", err)
	}
	fmt.Printf("Repo tool:    %s\n", repoPath)

	fmt.Println("Initializing repo...")
	repoInit := exec.Command(repoPath, "init",
		"-u", cfg.Manifest.URL,
		"-b", revision,
	)
	repoInit.Stdout = os.Stdout
	repoInit.Stderr = os.Stderr
	if err := repoInit.Run(); err != nil {
		return fmt.Errorf("repo init failed: %w", err)
	}

	if err := cfg.Cache(); err != nil {
		return fmt.Errorf("failed to cache config: %w", err)
	}

	fmt.Println("Init complete. Run 'ak yocto sync' next.")
	return nil
}

func ensureRepoTool() (string, error) {
	// Check if repo is already in PATH
	if path, err := exec.LookPath("repo"); err == nil {
		fmt.Println("Found repo tool in PATH.")
		return path, nil
	}

	// Check if we already downloaded it
	localRepo := filepath.Join(".ak", "bin", "repo")
	if info, err := os.Stat(localRepo); err == nil && !info.IsDir() {
		fmt.Println("Found previously downloaded repo tool.")
		return localRepo, nil
	}

	fmt.Println("Downloading Google repo tool...")

	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return "", fmt.Errorf("unsupported OS %s for repo tool", runtime.GOOS)
	}

	if err := os.MkdirAll(filepath.Join(".ak", "bin"), 0755); err != nil {
		return "", err
	}

	resp, err := http.Get(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to download repo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download repo: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(localRepo)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", err
	}

	if err := os.Chmod(localRepo, 0755); err != nil {
		return "", err
	}

	fmt.Println("Repo tool downloaded successfully.")
	return localRepo, nil
}
