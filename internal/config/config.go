package config

import (
	"os"

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
			if err := yaml.Unmarshal(data, &pc); err == nil {
				mergeConfig(&cfg, &pc.Default)
				if profile != "" && profile != "default" {
					if p, ok := pc.Profiles[profile]; ok {
						mergeConfig(&cfg, &p)
					}
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

func FileExists() bool {
	_, err := os.Stat(ConfigPath())
	return err == nil
}

func ResolveConfigPathForTests(dir string) string {
	return dir + "/config.yaml"
}
