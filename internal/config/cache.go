package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// DefaultListCacheTTL is how long list/completion cache entries stay fresh.
const DefaultListCacheTTL = 20 * time.Minute

// ClearCache removes cache files for a single profile only.
// Never deletes the cache root (which holds sibling profile caches).
func ClearCache(profile string) (bool, error) {
	target := ProfileCacheDir(profile)

	entries, err := os.ReadDir(target)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if len(entries) == 0 {
		return false, nil
	}
	for _, e := range entries {
		if err := os.RemoveAll(filepath.Join(target, e.Name())); err != nil {
			return false, err
		}
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

type cacheEnvelope struct {
	FetchedAt time.Time       `json:"fetched_at"`
	Data      json.RawMessage `json:"data"`
}

// WriteTTL stores data as JSON with a fetch timestamp under the profile cache dir.
func WriteTTL(profile, filename string, data any) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	env := cacheEnvelope{
		FetchedAt: time.Now().UTC(),
		Data:      raw,
	}
	b, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return WriteCache(profile, filename, b)
}

// ReadTTL loads a TTL cache entry into out. Returns false when missing or stale.
func ReadTTL(profile, filename string, out any, ttl time.Duration) (bool, error) {
	raw, err := ReadCache(profile, filename)
	if err != nil {
		return false, nil
	}
	var env cacheEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return false, nil
	}
	if env.FetchedAt.IsZero() {
		return false, nil
	}
	if ttl > 0 && time.Since(env.FetchedAt) > ttl {
		return false, nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return false, nil
	}
	return true, nil
}

// CacheTTLForEnv overrides DefaultListCacheTTL via CLOCKIFY_CACHE_TTL (Go duration).
func CacheTTLForEnv() time.Duration {
	if v := os.Getenv("CLOCKIFY_CACHE_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return DefaultListCacheTTL
}
