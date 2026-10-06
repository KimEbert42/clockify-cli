package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Config struct {
	APIKey      string `yaml:"api_key" mapstructure:"api_key"`
	WorkspaceID string `yaml:"workspace_id" mapstructure:"workspace_id"`
	Region      string `yaml:"region" mapstructure:"region"`
	Output      string `yaml:"output" mapstructure:"output"`
	DateFormat  string `yaml:"date_format" mapstructure:"date_format"`
	TimeFormat  string `yaml:"time_format" mapstructure:"time_format"`
	WeekStarts  string `yaml:"week_starts" mapstructure:"week_starts"`
	// AutoUpdate: empty = enabled (default); "false"/"0"/"no"/"off" disables passive checks.
	AutoUpdate string `yaml:"auto_update,omitempty" mapstructure:"auto_update"`
	// Interactive: empty = off; "true"/"1"/"yes"/"on" enables opt-in pickers on TTY.
	Interactive string `yaml:"interactive,omitempty" mapstructure:"interactive"`
	// TimeZone: IANA name (e.g. "America/New_York"); empty = OS local.
	TimeZone string `yaml:"timezone,omitempty" mapstructure:"timezone"`
}

type ProfileConfig struct {
	Version  int               `yaml:"version"`
	Default  Config            `yaml:"default"`
	Profiles map[string]Config `yaml:"profiles,omitempty"`
}

var defaults = Config{
	Region:     "global",
	Output:     "",
	DateFormat: "2006-01-02",
	TimeFormat: "15:04",
	WeekStarts: "MONDAY",
}

func Load(profile string) (*Config, error) {
	cfg := defaults

	if _, err := os.Stat(ConfigPath()); err == nil {
		data, err := os.ReadFile(ConfigPath())
		if err == nil {
			var pc ProfileConfig
			if err := yaml.Unmarshal(data, &pc); err != nil {
				return nil, fmt.Errorf("corrupt config %s: %w (fix or delete the file, then run 'clockify configure')", ConfigPath(), err)
			}
			mergeConfig(&cfg, &pc.Default)
			if profile != "" && profile != "default" {
				if p, ok := pc.Profiles[profile]; ok {
					mergeConfig(&cfg, &p)
				}
			}
		}
	}

	applyEnvOverrides(&cfg)

	return &cfg, nil
}

func mergeConfig(base, override *Config) {
	if override.APIKey != "" {
		base.APIKey = override.APIKey
	}
	if override.WorkspaceID != "" {
		base.WorkspaceID = override.WorkspaceID
	}
	if override.Region != "" {
		base.Region = override.Region
	}
	if override.Output != "" {
		base.Output = override.Output
	}
	if override.DateFormat != "" {
		base.DateFormat = override.DateFormat
	}
	if override.TimeFormat != "" {
		base.TimeFormat = override.TimeFormat
	}
	if override.WeekStarts != "" {
		base.WeekStarts = override.WeekStarts
	}
	if override.AutoUpdate != "" {
		base.AutoUpdate = override.AutoUpdate
	}
	if override.Interactive != "" {
		base.Interactive = override.Interactive
	}
	if override.TimeZone != "" {
		base.TimeZone = override.TimeZone
	}
}

// InteractiveEnabled reports whether interactive pickers are enabled via config.
func (c *Config) InteractiveEnabled() bool {
	if c == nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(c.Interactive)) {
	case "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}

// ParseInteractive normalizes a user-provided interactive value.
func ParseInteractive(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes", "on":
		return "true", nil
	case "false", "0", "no", "off":
		return "false", nil
	default:
		return "", fmt.Errorf("interactive must be true or false, got %q", value)
	}
}

// AutoUpdateEnabled reports whether passive update checks are enabled.
// Empty means enabled (default).
func (c *Config) AutoUpdateEnabled() bool {
	if c == nil {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(c.AutoUpdate)) {
	case "false", "0", "no", "off":
		return false
	default:
		return true
	}
}

// ParseAutoUpdate normalizes a user-provided auto_update value.
func ParseAutoUpdate(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes", "on":
		return "true", nil
	case "false", "0", "no", "off":
		return "false", nil
	default:
		return "", fmt.Errorf("auto_update must be true or false, got %q", value)
	}
}

func applyEnvOverrides(cfg *Config) {
	v := viper.New()
	v.SetEnvPrefix("CLOCKIFY")
	v.AutomaticEnv()
	v.BindEnv("api_key", "CLOCKIFY_API_KEY")
	v.BindEnv("workspace_id", "CLOCKIFY_WORKSPACE_ID")
	v.BindEnv("region", "CLOCKIFY_REGION")
	v.BindEnv("output", "CLOCKIFY_OUTPUT")

	if v.GetString("api_key") != "" {
		cfg.APIKey = v.GetString("api_key")
	}
	if v.GetString("workspace_id") != "" {
		cfg.WorkspaceID = v.GetString("workspace_id")
	}
	if v.GetString("region") != "" {
		cfg.Region = v.GetString("region")
	}
	if v.GetString("output") != "" {
		cfg.Output = v.GetString("output")
	}
}

func Save(cfg *ProfileConfig) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(), data, 0600)
}

func MergeFlags(cfg *Config, flags map[string]interface{}) *Config {
	result := *cfg
	if v, ok := flags["api-key"].(string); ok && v != "" {
		result.APIKey = v
	}
	if v, ok := flags["workspace"].(string); ok && v != "" {
		result.WorkspaceID = v
	}
	if v, ok := flags["region"].(string); ok && v != "" {
		result.Region = v
	}
	if v, ok := flags["output"].(string); ok && v != "" {
		result.Output = v
	}
	return &result
}

func LoadRaw() (*ProfileConfig, error) {
	if _, err := os.Stat(ConfigPath()); err != nil {
		return &ProfileConfig{Version: 1, Default: defaults, Profiles: make(map[string]Config)}, nil
	}
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	var pc ProfileConfig
	if err := yaml.Unmarshal(data, &pc); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	if pc.Profiles == nil {
		pc.Profiles = make(map[string]Config)
	}
	return &pc, nil
}

func LoadProfileNames() ([]string, error) {
	pc, err := LoadRaw()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(pc.Profiles)+1)
	names = append(names, "default")
	for name := range pc.Profiles {
		names = append(names, name)
	}
	return names, nil
}

func DeleteProfile(name string) error {
	if name == "default" {
		return fmt.Errorf("cannot delete the default profile")
	}
	pc, err := LoadRaw()
	if err != nil {
		return err
	}
	if _, ok := pc.Profiles[name]; !ok {
		return fmt.Errorf("profile %q does not exist", name)
	}
	delete(pc.Profiles, name)
	return Save(pc)
}

func FileExists() bool {
	_, err := os.Stat(ConfigPath())
	return err == nil
}

func ResolveConfigPathForTests(dir string) string {
	return dir + "/config.yaml"
}
