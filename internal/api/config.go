package api

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Bind string `yaml:"bind"`
	} `yaml:"server"`

	Auth struct {
		Token string `yaml:"token"`
	} `yaml:"auth"`

	Caddy struct {
		AdminSocket string `yaml:"adminSocket"`
		Caddyfile   string `yaml:"caddyfile"`
		SitesDir    string `yaml:"sitesDir"`
		RulesRoot   string `yaml:"rulesRoot"`
	} `yaml:"caddy"`

	Backup BackupConfig `yaml:"backup"`
}

type BackupConfig struct {
	Enabled bool   `yaml:"enabled"`
	Daily   string `yaml:"daily"`
	S3      struct {
		Endpoint  string `yaml:"endpoint"`
		Region    string `yaml:"region"`
		Bucket    string `yaml:"bucket"`
		AccessKey string `yaml:"accessKey"`
		SecretKey string `yaml:"secretKey"`
		Prefix    string `yaml:"prefix"`
	} `yaml:"s3"`
}

type CaddyConfig struct {
	AdminSocket string `yaml:"adminSocket"`
	Caddyfile   string `yaml:"caddyfile"`
	SitesDir    string `yaml:"sitesDir"`
	RulesRoot   string `yaml:"rulesRoot"`
}

func LoadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	if cfg.Server.Bind == "" {
		cfg.Server.Bind = ":8080"
	}
	return &cfg, nil
}
