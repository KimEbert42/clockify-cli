package config

import (
	"os"
	"path/filepath"
)

const (
	configDirName  = ".clockify"
	configFileName = "config.yaml"
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
