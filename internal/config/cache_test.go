package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCacheDir(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CACHE_HOME", dir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	got := CacheDir()
	want := filepath.Join(dir, "clockify")
	if got != want {
		t.Errorf("CacheDir() = %q, want %q", got, want)
	}
}

func TestProfileCacheDir(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CACHE_HOME", dir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	defaultDir := ProfileCacheDir("default")
	wantDefault := filepath.Join(dir, "clockify")
	if defaultDir != wantDefault {
		t.Errorf("ProfileCacheDir('default') = %q, want %q", defaultDir, wantDefault)
	}

	emptyDir := ProfileCacheDir("")
	if emptyDir != wantDefault {
		t.Errorf("ProfileCacheDir('') = %q, want %q", emptyDir, wantDefault)
	}

	workDir := ProfileCacheDir("work")
	wantWork := filepath.Join(dir, "clockify", "work")
	if workDir != wantWork {
		t.Errorf("ProfileCacheDir('work') = %q, want %q", workDir, wantWork)
	}
}

func TestClearCache(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CACHE_HOME", dir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	cacheDir := filepath.Join(dir, "clockify")
	os.MkdirAll(cacheDir, 0700)
	os.WriteFile(filepath.Join(cacheDir, "test.json"), []byte(`{}`), 0600)

	cleared, err := ClearCache("default")
	if err != nil {
		t.Fatalf("ClearCache() error = %v", err)
	}
	if !cleared {
		t.Error("ClearCache() should return true when files exist")
	}

	if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
		t.Error("cache directory should have been removed")
	}
}

func TestClearCacheNonexistent(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CACHE_HOME", dir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	cleared, err := ClearCache("default")
	if err != nil {
		t.Fatalf("ClearCache() error = %v", err)
	}
	if cleared {
		t.Error("ClearCache() should return false when no cache exists")
	}
}

func TestClearCacheProfile(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CACHE_HOME", dir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	defaultCache := filepath.Join(dir, "clockify")
	workCache := filepath.Join(dir, "clockify", "work")
	os.MkdirAll(defaultCache, 0700)
	os.MkdirAll(workCache, 0700)
	os.WriteFile(filepath.Join(defaultCache, "default.json"), []byte(`{}`), 0600)
	os.WriteFile(filepath.Join(workCache, "work.json"), []byte(`{}`), 0600)

	cleared, err := ClearCache("work")
	if err != nil {
		t.Fatalf("ClearCache('work') error = %v", err)
	}
	if !cleared {
		t.Error("ClearCache('work') should return true")
	}

	if _, err := os.Stat(defaultCache); os.IsNotExist(err) {
		t.Error("default cache should not have been removed")
	}
	if _, err := os.Stat(workCache); !os.IsNotExist(err) {
		t.Error("work cache should have been removed")
	}
}

func TestWriteAndReadCache(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CACHE_HOME", dir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	data := []byte(`{"projects": []}`)
	if err := WriteCache("default", "projects.json", data); err != nil {
		t.Fatalf("WriteCache() error = %v", err)
	}

	got, err := ReadCache("default", "projects.json")
	if err != nil {
		t.Fatalf("ReadCache() error = %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("ReadCache() = %q, want %q", got, data)
	}
}

func TestCacheExists(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CACHE_HOME", dir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	if CacheExists("default", "test.json") {
		t.Error("CacheExists() should return false for non-existent file")
	}

	WriteCache("default", "test.json", []byte(`{}`))

	if !CacheExists("default", "test.json") {
		t.Error("CacheExists() should return true after writing")
	}
}
