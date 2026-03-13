package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ramaxlo/akcli/config"
	"github.com/spf13/cobra"
)

var packCmd = &cobra.Command{
	Use:   "pack",
	Short: "Pack built image artifacts into a tar.gz file",
	RunE:  runPack,
}

func init() {
	yoctoCmd.AddCommand(packCmd)
}

func runPack(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadCache()
	if err != nil {
		return err
	}

	machine := cfg.Build.Machine
	imageDir := filepath.Join(cfg.Build.BuildDir, "tmp", "deploy", "images", machine)

	if info, err := os.Stat(imageDir); err != nil || !info.IsDir() {
		return fmt.Errorf("image directory not found: %s", imageDir)
	}

	output := fmt.Sprintf("%s.tar.gz", machine)

	if dryRun {
		fmt.Printf("[dryrun] Would pack %s into %s with prefix %s/\n", imageDir, output, machine)
		return nil
	}

	fmt.Printf("Packing artifacts from %s...\n", imageDir)

	absOutput, err := filepath.Abs(output)
	if err != nil {
		return fmt.Errorf("failed to resolve output path: %w", err)
	}

	// Use --transform with flags=rSh to only rename regular files, symlinks
	// (path entries), and hard links, but NOT symlink targets
	transform := fmt.Sprintf("--transform=flags=rsh;s,^\\.,%s,", machine)
	tarCmd := exec.Command("tar", "-czf", absOutput, transform, "-C", imageDir, ".")
	tarCmd.Stdout = os.Stdout
	tarCmd.Stderr = os.Stderr
	if err := tarCmd.Run(); err != nil {
		return fmt.Errorf("failed to create tarball: %w", err)
	}

	fmt.Printf("Created %s\n", output)
	return nil
}
