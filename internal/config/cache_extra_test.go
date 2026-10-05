package config

import (
	"os"
	"testing"
	"time"
)

func TestWriteReadCacheRoundTrip(t *testing.T) {
	withTempConfigHome(t)
	if err := WriteCache("default", "list.json", []byte(`{"ok":true}`)); err != nil {
		t.Fatal(err)
	}
	if !CacheExists("default", "list.json") {
		t.Fatal("cache should exist")
	}
	data, err := ReadCache("default", "list.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"ok":true}` {
		t.Errorf("data=%q", data)
	}
	if _, err := ReadCache("default", "missing.json"); err == nil {
		t.Error("expected missing cache error")
	}
}

func TestWriteReadTTL(t *testing.T) {
	withTempConfigHome(t)
	type payload struct {
		N int `json:"n"`
	}
	if err := WriteTTL("default", "ttl.json", payload{N: 42}); err != nil {
		t.Fatal(err)
	}
	var out payload
	ok, err := ReadTTL("default", "ttl.json", &out, time.Minute)
	if err != nil || !ok || out.N != 42 {
		t.Fatalf("ok=%v err=%v out=%+v", ok, err, out)
	}

	// Stale entry
	envPath := CacheFilePath("default", "ttl.json")
	raw, _ := os.ReadFile(envPath)
	// Rewrite with old timestamp
	stale := []byte(`{"fetched_at":"2020-01-01T00:00:00Z","data":{"n":1}}`)
	_ = os.WriteFile(envPath, stale, 0600)
	ok, err = ReadTTL("default", "ttl.json", &out, time.Minute)
	if ok || err != nil {
		t.Fatalf("stale should miss: ok=%v err=%v", ok, err)
	}

	// Zero TTL means no expiry check? Code: `if ttl > 0 && ...` so 0 = always fresh if present
	_ = os.WriteFile(envPath, stale, 0600)
	ok, _ = ReadTTL("default", "ttl.json", &out, 0)
	if !ok {
		t.Fatal("ttl=0 should accept any fetched_at")
	}

	// Invalid envelope
	_ = os.WriteFile(envPath, []byte(`not-json`), 0600)
	ok, _ = ReadTTL("default", "ttl.json", &out, time.Minute)
	if ok {
		t.Fatal("invalid envelope should miss")
	}

	// Zero FetchedAt
	_ = os.WriteFile(envPath, []byte(`{"fetched_at":"0001-01-01T00:00:00Z","data":{"n":9}}`), 0600)
	ok, _ = ReadTTL("default", "ttl.json", &out, time.Minute)
	if ok {
		t.Fatal("zero fetched_at should miss")
	}

	// Missing file
	ok, err = ReadTTL("default", "nope.json", &out, time.Minute)
	if ok || err != nil {
		t.Fatalf("missing: ok=%v err=%v", ok, err)
	}
	_ = raw
}

func TestCacheTTLForEnv(t *testing.T) {
	t.Setenv("CLOCKIFY_CACHE_TTL", "")
	if got := CacheTTLForEnv(); got != DefaultListCacheTTL {
		t.Errorf("default=%v", got)
	}
	t.Setenv("CLOCKIFY_CACHE_TTL", "5m")
	if got := CacheTTLForEnv(); got != 5*time.Minute {
		t.Errorf("5m=%v", got)
	}
	t.Setenv("CLOCKIFY_CACHE_TTL", "not-a-duration")
	if got := CacheTTLForEnv(); got != DefaultListCacheTTL {
		t.Errorf("invalid=%v", got)
	}
	t.Setenv("CLOCKIFY_CACHE_TTL", "-3s")
	if got := CacheTTLForEnv(); got != DefaultListCacheTTL {
		t.Errorf("negative=%v", got)
	}
}

func TestClearCacheAndCacheFilePath(t *testing.T) {
	withTempConfigHome(t)
	if cleared, err := ClearCache("default"); err != nil || cleared {
		t.Fatalf("clear empty: cleared=%v err=%v", cleared, err)
	}
	if err := WriteCache("work", "x.json", []byte("1")); err != nil {
		t.Fatal(err)
	}
	cleared, err := ClearCache("work")
	if err != nil || !cleared {
		t.Fatalf("clear: cleared=%v err=%v", cleared, err)
	}
	if CacheExists("work", "x.json") {
		t.Error("cache still exists after clear")
	}
	if CacheFilePath("work", "x.json") == "" {
		t.Error("empty cache path")
	}
}
