package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type ManifestConfig struct {
	URL    string `toml:"url"`
	Branch string `toml:"branch"`
	Tag    string `toml:"tag"`
}

type BuildConfig struct {
	Distro       string `toml:"distro"`
	Machine      string `toml:"machine"`
	Target       string `toml:"target"`
	BuildDir     string `toml:"build_dir"`
	TemplateConf string `toml:"template_conf"`
	DlDir        string `toml:"dl_dir"`
	SstateDir    string `toml:"sstate_dir"`
}

type KernelRemote struct {
	Name     string `toml:"name"`
	URL      string `toml:"url"`
	Branch   string `toml:"branch"`
	Checkout bool   `toml:"checkout"`
}

type KernelDefconfig struct {
	Name   string `toml:"name"`
	Config string `toml:"config"`
}

type KernelConfig struct {
	SrcDir          string            `toml:"src_dir"`
	Arch            string            `toml:"arch"`
	ToolchainPrefix string            `toml:"toolchain_prefix"`
	Targets         []string          `toml:"targets"`
	Remotes         []KernelRemote    `toml:"remote"`
	Defconfigs      []KernelDefconfig `toml:"defconfig"`
}

type Config struct {
	Manifest ManifestConfig `toml:"manifest"`
	Build    BuildConfig    `toml:"build"`
	Kernel   KernelConfig   `toml:"kernel"`
}

const cacheDir = ".ak"
const cacheFile = "config.cache.toml"

func Load(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config %s: %w", path, err)
	}

	if cfg.Build.BuildDir == "" {
		cfg.Build.BuildDir = "build"
	}

	if cfg.Manifest.Branch != "" && cfg.Manifest.Tag != "" {
		return nil, fmt.Errorf("manifest: 'branch' and 'tag' are mutually exclusive")
	}

	return &cfg, nil
}

func (c *Config) Cache() error {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache dir: %w", err)
	}

	f, err := os.Create(filepath.Join(cacheDir, cacheFile))
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer f.Close()

	if err := toml.NewEncoder(f).Encode(c); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

func LoadCache() (*Config, error) {
	path := filepath.Join(cacheDir, cacheFile)
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to load cached config (did you run 'ak yocto init'?): %w", err)
	}
	return &cfg, nil
}

const kernelCacheFile = "kernel.cache.toml"

func (c *Config) KernelCache() error {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache dir: %w", err)
	}

	f, err := os.Create(filepath.Join(cacheDir, kernelCacheFile))
	if err != nil {
		return fmt.Errorf("failed to create kernel cache file: %w", err)
	}
	defer f.Close()

	if err := toml.NewEncoder(f).Encode(c); err != nil {
		return fmt.Errorf("failed to write kernel cache: %w", err)
	}

	return nil
}

func LoadKernelCache() (*Config, error) {
	path := filepath.Join(cacheDir, kernelCacheFile)
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to load kernel cache (did you run 'ak kernel init'?): %w", err)
	}
	return &cfg, nil
}
