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
	// RequiredTags is the list of tags that must be present on a DN host in
	// the network in order for a DNS record to be created for it.
	RequiredTags []string `toml:"required_tags"`
	// RequiredSuffix is the suffix that must be present in the DN hostname in
	// order for a DNS record to be created for it.
	RequiredSuffix string `toml:"required_suffix"`

	// TrimSuffix determines whether to trim everything after the first . in
	// the DN hostname before creating the DNS record.
	TrimSuffix bool `toml:"trim_suffix"`
	// AppendSuffix is the suffix to append to the hostname from DN before
	// creating the DNS record (occurs after TrimSuffix.)
	AppendSuffix string `toml:"append_suffix"`

	// PruneRecords indicates whether to delete Cloudflare DNS records that
	// weren't created this run.
	PruneRecords bool `toml:"prune_records"`
	// PruneNetworkRecordsOnly indicates whether to only delete records that
	// match the network CIDR from Defined Networking.
	PruneNetworkRecordsOnly bool `toml:"prune_network_records_only"`

	// Cloudflare is the configuration for the Cloudflare API.
	Cloudflare CloudflareConfig `toml:"cloudflare"`

	// Defined is the configuration for the Defined Networking API.
	DefinedNet DefinedConfig `toml:"definednet"`
}

type CloudflareConfig struct {
	APIToken string `toml:"api_token"`
	// ZoneName is the name of the Cloudflare zone to manage DNS records in.
	ZoneName string `toml:"zone_name"`
}

type DefinedConfig struct {
	APIToken string `toml:"api_token"`
	// NetworkID is the ID of the network in Defined Networking to monitor for
	// hosts.
	NetworkID string `toml:"network_id"`
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
