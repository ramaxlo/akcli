package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
)

var syncJobs int

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Fetch yocto meta layers using repo sync",
	RunE:  runSync,
}

func init() {
	syncCmd.Flags().IntVarP(&syncJobs, "jobs", "j", 4, "number of parallel sync jobs")
	yoctoCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	if dryRun {
		fmt.Printf("[dryrun] Would run: repo sync -j %d\n", syncJobs)
		return nil
	}

	repoPath, err := findRepo()
	if err != nil {
		return err
	}

	fmt.Printf("Syncing with %d jobs...\n", syncJobs)
	fmt.Printf("Running: %s sync -j %d\n", repoPath, syncJobs)

	repoSync := exec.Command(repoPath, "sync", "-j", strconv.Itoa(syncJobs))
	repoSync.Stdout = os.Stdout
	repoSync.Stderr = os.Stderr
	if err := repoSync.Run(); err != nil {
		return fmt.Errorf("repo sync failed: %w", err)
	}

	fmt.Println("Sync complete. Run 'ak yocto setup' next.")
	return nil
}

func findRepo() (string, error) {
	// Check local download first
	localRepo := filepath.Join(".ak", "bin", "repo")
	if info, err := os.Stat(localRepo); err == nil && !info.IsDir() {
		return localRepo, nil
	}

	// Check PATH
	if path, err := exec.LookPath("repo"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("repo tool not found (did you run 'ak yocto init'?)")
}
