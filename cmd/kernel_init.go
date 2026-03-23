package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/ramaxlo/akcli/config"
	"github.com/spf13/cobra"
)

var kernelInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize kernel git repository and fetch remotes",
	RunE:  runKernelInit,
}

func init() {
	kernelCmd.AddCommand(kernelInitCmd)
}

func runKernelInit(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	k := cfg.Kernel

	if k.SrcDir == "" {
		return fmt.Errorf("kernel.src_dir is required")
	}
	if k.ToolchainPrefix == "" {
		return fmt.Errorf("kernel.toolchain_prefix is required")
	}
	if len(k.Remotes) == 0 {
		return fmt.Errorf("at least one [[kernel.remote]] is required")
	}

	// Find the last remote with checkout=true
	var checkoutIdx int = -1
	for i := range k.Remotes {
		if k.Remotes[i].Checkout {
			checkoutIdx = i
		}
	}

	if dryRun {
		fmt.Printf("[dryrun] Would run: git init %s\n", k.SrcDir)
		for _, r := range k.Remotes {
			fmt.Printf("[dryrun] Would run: git -C %s remote add %s %s\n", k.SrcDir, r.Name, r.URL)
			fmt.Printf("[dryrun] Would run: git -C %s fetch %s\n", k.SrcDir, r.Name)
		}
		if checkoutIdx >= 0 {
			r := k.Remotes[checkoutIdx]
			fmt.Printf("[dryrun] Would run: git -C %s checkout %s/%s\n", k.SrcDir, r.Name, r.Branch)
		}
		fmt.Printf("[dryrun] Would verify toolchain: %sgcc in PATH\n", k.ToolchainPrefix)
		fmt.Println("[dryrun] Would cache kernel config to .ak/kernel.cache.toml")
		return nil
	}

	if err := runCmd("git", "init", k.SrcDir); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}

	for _, r := range k.Remotes {
		if err := runCmd("git", "-C", k.SrcDir, "remote", "add", r.Name, r.URL); err != nil {
			return fmt.Errorf("git remote add %s failed: %w", r.Name, err)
		}
		if err := runCmd("git", "-C", k.SrcDir, "fetch", r.Name); err != nil {
			return fmt.Errorf("git fetch %s failed: %w", r.Name, err)
		}
	}

	if checkoutIdx >= 0 {
		r := k.Remotes[checkoutIdx]
		ref := fmt.Sprintf("%s/%s", r.Name, r.Branch)
		if err := runCmd("git", "-C", k.SrcDir, "checkout", ref); err != nil {
			return fmt.Errorf("git checkout %s failed: %w", ref, err)
		}
	}

	gccBinary := k.ToolchainPrefix + "gcc"
	gccPath, err := exec.LookPath(gccBinary)
	if err != nil {
		return fmt.Errorf("toolchain not found: %s not in PATH", gccBinary)
	}
	fmt.Printf("Toolchain found: %s\n", filepath.Dir(gccPath))

	if err := cfg.KernelCache(); err != nil {
		return fmt.Errorf("failed to cache kernel config: %w", err)
	}

	fmt.Println("Kernel init complete. Run 'ak kernel build' next.")
	return nil
}
