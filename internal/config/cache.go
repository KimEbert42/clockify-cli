package config

import (
	"os"
	"path/filepath"
)

func ClearCache(profile string) (bool, error) {
	target := ProfileCacheDir(profile)

	if _, err := os.Stat(target); os.IsNotExist(err) {
		return false, nil
	}

	if err := os.RemoveAll(target); err != nil {
		return false, err
	}

	return true, nil
}

func CacheFilePath(profile, filename string) string {
	return filepath.Join(ProfileCacheDir(profile), filename)
}

func WriteCache(profile, filename string, data []byte) error {
	if err := EnsureCacheDir(profile); err != nil {
		return err
	}
	return os.WriteFile(CacheFilePath(profile, filename), data, 0600)
}

func ReadCache(profile, filename string) ([]byte, error) {
	return os.ReadFile(CacheFilePath(profile, filename))
}

func CacheExists(profile, filename string) bool {
	_, err := os.Stat(CacheFilePath(profile, filename))
	return err == nil
}
