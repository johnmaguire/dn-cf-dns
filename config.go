package main

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/johnmaguire/nebula-dns/internal/envconfig"
)

type AppConfig struct {
	// RequiredTags is the list of tags that must be present on a DN host in
	// the network in order for a DNS record to be created for it.
	RequiredTags []string `toml:"required_tags" envconfig:"NEBULA_DNS_REQUIRED_TAGS"`
	// RequiredSuffix is the suffix that must be present in the DN hostname in
	// order for a DNS record to be created for it.
	RequiredSuffix string `toml:"required_suffix" envconfig:"NEBULA_DNS_REQUIRED_SUFFIX"`

	// TrimSuffix determines whether to trim everything after the first . in
	// the DN hostname before creating the DNS record.
	TrimSuffix bool `toml:"trim_suffix" envconfig:"NEBULA_DNS_TRIM_SUFFIX"`
	// AppendSuffix is the suffix to append to the hostname from DN before
	// creating the DNS record (occurs after TrimSuffix.)
	AppendSuffix string `toml:"append_suffix" envconfig:"NEBULA_DNS_APPEND_SUFFIX"`

	// PruneRecords indicates whether to delete Cloudflare DNS records that
	// weren't created this run.
	PruneRecords bool `toml:"prune_records" envconfig:"NEBULA_DNS_PRUNE_RECORDS"`
	// PruneNetworkRecordsOnly indicates whether to only delete records that
	// match the network CIDR from Defined Networking.
	PruneNetworkRecordsOnly bool `toml:"prune_network_records_only" envconfig:"NEBULA_DNS_PRUNE_NETWORK_RECORDS_ONLY"`

	// Cloudflare is the configuration for the Cloudflare API.
	Cloudflare CloudflareConfig `toml:"cloudflare"`

	// Defined is the configuration for the Defined Networking API.
	DefinedNet DefinedConfig `toml:"definednet"`
}

type CloudflareConfig struct {
	APIToken string `toml:"api_token" envconfig:"NEBULA_DNS_CF_API_TOKEN"`
	// ZoneName is the name of the Cloudflare zone to manage DNS records in.
	ZoneName string `toml:"zone_name" envconfig:"NEBULA_DNS_CF_ZONE_NAME"`
}

type DefinedConfig struct {
	APIToken string `toml:"api_token" envconfig:"NEBULA_DNS_DN_API_TOKEN"`
	// NetworkID is the ID of the network in Defined Networking to monitor for
	// hosts.
	NetworkID string `toml:"network_id" envconfig:"NEBULA_DNS_DN_NETWORK_ID"`
}

func LoadConfig(path string) (*AppConfig, error) {
	// Load config from file if it exists
	config, err := newConfigFromFile(path)
	if err != nil {
		return nil, err
	}

	// Overlay environment variables onto config
	if err := envconfig.Process(config); err != nil {
		return nil, err
	}

	// Validate required fields
	if err := config.validate(); err != nil {
		return nil, err
	}

	// Default AppendSuffix to the zone name
	if config.AppendSuffix == "" {
		config.AppendSuffix = config.Cloudflare.ZoneName
	}

	return config, nil
}

func (c *AppConfig) validate() error {
	var missing []string
	if c.Cloudflare.APIToken == "" {
		missing = append(missing, "cloudflare.api_token")
	}
	if c.Cloudflare.ZoneName == "" {
		missing = append(missing, "cloudflare.zone_name")
	}
	if c.DefinedNet.APIToken == "" {
		missing = append(missing, "definednet.api_token")
	}
	if c.DefinedNet.NetworkID == "" {
		missing = append(missing, "definednet.network_id")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required config fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

func newConfigFromFile(path string) (*AppConfig, error) {
	var config AppConfig
	if path == "" {
		return &config, nil
	}
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
