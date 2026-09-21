package config

import (
	"os"
	"testing"

	"github.com/zalando/go-keyring"
)

func init() {
	keyring.MockInit()
}

func TestGetAPIKey_EnvVar(t *testing.T) {
	os.Setenv("CLOCKIFY_API_KEY", "test-api-key")
	defer os.Unsetenv("CLOCKIFY_API_KEY")

	key, err := GetAPIKey("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "test-api-key" {
		t.Errorf("expected 'test-api-key', got '%s'", key)
	}
}

func TestGetAPIKey_EnvVarPriority(t *testing.T) {
	os.Setenv("CLOCKIFY_API_KEY", "env-key")
	defer os.Unsetenv("CLOCKIFY_API_KEY")

	key, err := GetAPIKey("default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "env-key" {
		t.Errorf("expected 'env-key', got '%s'", key)
	}
}

func TestGetAPIKey_NoKey(t *testing.T) {
	os.Unsetenv("CLOCKIFY_API_KEY")

	_, err := GetAPIKey("")
	if err == nil {
		t.Error("expected error when no API key configured")
	}
}

func TestSetAPIKey_Keyring(t *testing.T) {
	os.Unsetenv("CLOCKIFY_API_KEY")

	err := SetAPIKey("", "new-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	key, err := getKeyring(keyringUser)
	if err != nil {
		t.Fatalf("failed to get key from keyring: %v", err)
	}
	if key != "new-key" {
		t.Errorf("expected 'new-key' in keyring, got '%s'", key)
	}
}

func TestDeleteAPIKey_Keyring(t *testing.T) {
	os.Unsetenv("CLOCKIFY_API_KEY")

	err := SetAPIKey("", "to-delete")
	if err != nil {
		t.Fatalf("failed to set key: %v", err)
	}

	err = DeleteAPIKey("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = getKeyring(keyringUser)
	if err == nil {
		t.Error("expected error when getting deleted key from keyring")
	}
}

func TestGetAPIKey_ProfileKey(t *testing.T) {
	os.Setenv("CLOCKIFY_API_KEY", "default-key")
	defer os.Unsetenv("CLOCKIFY_API_KEY")

	key, err := GetAPIKey("work")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "default-key" {
		t.Errorf("expected 'default-key', got '%s'", key)
	}
}
