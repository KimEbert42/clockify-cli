package config

import (
	"os"
	"path/filepath"
)

const (
	configDirName  = ".clockify"
	configFileName = "config.yaml"
	cacheDirName   = "cache"
)

func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "clockify")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, configDirName)
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), configFileName)
}

func EnsureConfigDir() error {
	return os.MkdirAll(ConfigDir(), 0700)
}

func CacheDir() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "clockify")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "clockify")
}

func ProfileCacheDir(profile string) string {
	base := CacheDir()
	if profile != "" && profile != "default" {
		return filepath.Join(base, profile)
	}
	return base
}

func EnsureCacheDir(profile string) error {
	return os.MkdirAll(ProfileCacheDir(profile), 0700)
}
