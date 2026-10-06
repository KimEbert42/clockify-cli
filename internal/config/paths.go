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
		return filepath.Join(base, "profiles", profile)
	}
	// Default profile uses a dedicated subdir so clear-cache never wipes
	// sibling named-profile caches living under the cache root.
	return filepath.Join(base, "profiles", "default")
}

// DefaultProfileCacheDir returns the cache directory for the default profile.
func DefaultProfileCacheDir() string {
	return ProfileCacheDir("default")
}

// ProfileCacheRoot returns the parent directory that contains all profile caches.
func ProfileCacheRoot() string {
	return filepath.Join(CacheDir(), "profiles")
}

func EnsureCacheDir(profile string) error {
	return os.MkdirAll(ProfileCacheDir(profile), 0700)
}
