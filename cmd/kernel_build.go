package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/ramaxlo/akcli/config"
	"github.com/spf13/cobra"
)

var buildJobs int

var kernelBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the Linux kernel for all configured defconfigs",
	RunE:  runKernelBuild,
}

func init() {
	kernelBuildCmd.Flags().IntVarP(&buildJobs, "jobs", "j", runtime.NumCPU(), "number of parallel jobs for make")
	kernelCmd.AddCommand(kernelBuildCmd)
}

func logPrintf(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Print(msg)
	if kernelLogWriter != nil {
		kernelLogWriter.Write([]byte(msg))
	}
}

func runKernelBuild(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadKernelCache()
	if err != nil {
		return err
	}

	k := cfg.Kernel

	if k.BuildLog != "" && !dryRun {
		logFile, err := os.Create(k.BuildLog)
		if err != nil {
			return fmt.Errorf("failed to create build log %s: %w", k.BuildLog, err)
		}
		defer logFile.Close()

		kernelLogWriter = &timestampWriter{w: logFile}
		defer func() {
			kernelLogWriter.Flush()
			kernelLogWriter = nil
		}()
		fmt.Printf("Logging to: %s\n", k.BuildLog)
	}

	gccBinary := k.ToolchainPrefix + "gcc"
	if _, err := exec.LookPath(gccBinary); err != nil {
		return fmt.Errorf("toolchain not found: %s not in PATH", gccBinary)
	}

	jobs := strconv.Itoa(buildJobs)

	for _, d := range k.Defconfigs {
		outDir, err := filepath.Abs(filepath.Join("kbuild", d.Name))
		if err != nil {
			return fmt.Errorf("failed to resolve output dir for %s: %w", d.Name, err)
		}
		dotConfig := filepath.Join(outDir, ".config")

		if dryRun {
			fmt.Printf("[dryrun] Would create: %s\n", outDir)
			fmt.Printf("[dryrun] Would copy %s -> %s\n", d.Config, dotConfig)
			fmt.Printf("[dryrun] Would run: make -C %s ARCH=%s CROSS_COMPILE=%s O=%s olddefconfig\n",
				k.SrcDir, k.Arch, k.ToolchainPrefix, outDir)
			fmt.Printf("[dryrun] Would run: make -C %s ARCH=%s CROSS_COMPILE=%s O=%s -j%s %s\n",
				k.SrcDir, k.Arch, k.ToolchainPrefix, outDir, jobs, strings.Join(k.Targets, " "))
			continue
		}

		logPrintf("==> Remove old build folder: %s\n", outDir)
		// Remove old build folder
		if err := os.RemoveAll(outDir); err != nil {
			return fmt.Errorf("failed to remove old build folder: %w", err)
		}

		logPrintf("==> Defconfig: %s\n", d.Name)

		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create output dir %s: %w", outDir, err)
		}

		data, err := os.ReadFile(d.Config)
		if err != nil {
			return fmt.Errorf("failed to read config file %s: %w", d.Config, err)
		}
		if err := os.WriteFile(dotConfig, data, 0644); err != nil {
			return fmt.Errorf("failed to write .config for %s: %w", d.Name, err)
		}
		logPrintf("Copied %s -> %s\n", d.Config, dotConfig)

		if err := runMake(k.SrcDir, k.Arch, k.ToolchainPrefix, outDir, "olddefconfig"); err != nil {
			return fmt.Errorf("olddefconfig failed for %s: %w", d.Name, err)
		}

		for _, target := range k.Targets {
			buildTargets := append([]string{"-j" + jobs}, target)
			if err := runMake(k.SrcDir, k.Arch, k.ToolchainPrefix, outDir, buildTargets...); err != nil {
				return fmt.Errorf("build failed for %s: %w", d.Name, err)
			}
		}
	}

	if !dryRun {
		logPrintf("Kernel build complete.\n")
	}
	return nil
}

func runMake(srcDir, arch, crossCompile, outDir string, targets ...string) error {
	args := []string{
		"-C", srcDir,
		"ARCH=" + arch,
		"CROSS_COMPILE=" + crossCompile,
		"O=" + outDir,
	}
	args = append(args, targets...)
	return runCmd("make", args...)
}
