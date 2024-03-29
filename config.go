package main

import (
	"os"

	"github.com/BurntSushi/toml"
)

const (
	// CFAPIEnvVar is the name of the environment variable that contains the
	// Cloudflare API token.
	CFAPIEnvVar = "CF_API_TOKEN"

	// DNAPIEnvVar is the name of the environment variable that contains the
	// Defined Networking API token.
	DNAPIEnvVar = "DN_API_TOKEN"
)

type AppConfig struct {
	RequiredTags   []string `toml:"required_tags"`
	RequiredSuffix string   `toml:"required_suffix"`
	TrimSuffix     bool     `toml:"trim_suffix"`
	AppendSuffix   string   `toml:"append_suffix"`
	PruneRecords   bool     `toml:"prune_records"`

	// Cloudflare is the configuration for the Cloudflare API.
	Cloudflare CloudflareConfig `toml:"cloudflare"`

	// Defined is the configuration for the Defined Networking API.
	DefinedNet DefinedConfig `toml:"definednet"`
}

type CloudflareConfig struct {
	APIToken string `toml:"api_token"`
	ZoneName string `toml:"zone_name"`
}

type DefinedConfig struct {
	APIToken string `toml:"api_token"`
}

func LoadConfig(path string) (*AppConfig, error) {
	// Load config from file
	config, err := newConfigFromFile(path)
	if err != nil {
		return nil, err
	}

	// Optionally load secrets from the environment
	config.readEnv()

	// Default AppendSuffix to the zone name
	if config.AppendSuffix == "" {
		config.AppendSuffix = config.Cloudflare.ZoneName
	}

	return config, nil
}

func (c *AppConfig) readEnv() {
	if cfToken := os.Getenv(CFAPIEnvVar); cfToken != "" {
		c.Cloudflare.APIToken = cfToken
	}
	if dnToken := os.Getenv(DNAPIEnvVar); dnToken != "" {
		c.DefinedNet.APIToken = dnToken
	}
}

func newConfigFromFile(path string) (*AppConfig, error) {
	var config AppConfig
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
