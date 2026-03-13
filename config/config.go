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
}

type BuildConfig struct {
	Distro       string `toml:"distro"`
	Machine      string `toml:"machine"`
	Target       string `toml:"target"`
	BuildDir     string `toml:"build_dir"`
	TemplateConf string `toml:"template_conf"`
}

type Config struct {
	Manifest ManifestConfig `toml:"manifest"`
	Build    BuildConfig    `toml:"build"`
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
