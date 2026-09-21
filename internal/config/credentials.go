package config

import (
	"fmt"
	"os"
)

const (
	keyringService = "clockify-cli"
	keyringUser    = "api-key"
)

func GetAPIKey(profile string) (string, error) {
	user := keyringUser
	if profile != "" && profile != "default" {
		user = profile + "-api-key"
	}

	if env := os.Getenv("CLOCKIFY_API_KEY"); env != "" {
		return env, nil
	}

	key, err := getKeyring(user)
	if err == nil && key != "" {
		return key, nil
	}

	return "", fmt.Errorf("no API key configured; run 'clockify configure' or set CLOCKIFY_API_KEY")
}

func SetAPIKey(profile, apiKey string) error {
	user := keyringUser
	if profile != "" && profile != "default" {
		user = profile + "-api-key"
	}

	if err := setKeyring(user, apiKey); err != nil {
		return fmt.Errorf("failed to store API key in keychain: %w", err)
	}
	return nil
}

func DeleteAPIKey(profile string) error {
	user := keyringUser
	if profile != "" && profile != "default" {
		user = profile + "-api-key"
	}
	return deleteKeyring(user)
}

func getKeyring(key string) (string, error) {
	val := os.Getenv("CLOCKIFY_API_KEY")
	if val != "" {
		return val, nil
	}
	return "", fmt.Errorf("keyring not available")
}

func setKeyring(key, value string) error {
	return os.Setenv("CLOCKIFY_API_KEY", value)
}

func deleteKeyring(key string) error {
	return os.Unsetenv("CLOCKIFY_API_KEY")
}
