// nebula-dns is a tool for creating DNS records in Cloudflare based on hosts
// that exist in Defined Networking.
//
// This tool is expected to be run as a one-shot job periodically (e.g. once
// a minute) to ensure that DNS records are kept up to date.
//
// It is too bad that Defined Networking does not support webhooks or we would
// be able to trigger this tool on demand when hosts are created or destroyed.
package main

import (
	"context"
	"fmt"
	"net/mail"
	"net/netip"
	"os"
	"runtime/debug"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

// version can be set at build time via -ldflags "-X main.version=1.2.3".
// If not set, it falls back to the module version from Go's embedded build info
// (populated by go install module@version).
var version = ""

func init() {
	if version != "" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		version = info.Main.Version
	} else {
		version = "dev"
	}
}

func main() {
	if err := mainWithErr(); err != nil {
		log.Fatal().Err(err).Send()
	}
}

func mainWithErr() error {
	cmd := &cli.Command{
		Name:        "nebula-dns",
		Version:     version,
		Description: "nebula-dns manages DNS records in Cloudflare based on Defined Networking hosts",
		Authors: []any{
			&mail.Address{Name: "John Maguire", Address: "contact@johnmaguire.me"},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Usage: "path to config file"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			// Read the config file
			cfg, err := LoadConfig(c.String("config"))
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			cf, err := cloudflare.NewWithAPIToken(cfg.Cloudflare.APIToken)
			if err != nil {
				return err
			}

			// Find the Cloudflare zone ID for the zone we're interested in
			zoneID, err := GetZoneID(cf, cfg.Cloudflare.ZoneName)
			if err != nil {
				return fmt.Errorf("failed to get zone ID: %w", err)
			}
			log.Info().Str("zoneID", zoneID).Msgf("Found Cloudflare zone ID for %s", cfg.Cloudflare.ZoneName)

			// Get the network CIDR for the network we're interested in
			cidr, err := GetNetworkCIDR(cfg.DefinedNet.APIToken, cfg.DefinedNet.NetworkID)
			if err != nil {
				return fmt.Errorf("failed to get network CIDR: %w", err)
			}
			log.Info().Str("networkCIDR", cidr.String()).Msgf("Found network CIDR for network %s", cfg.DefinedNet.NetworkID)

			// Filter the DN hosts based on the following criteria:
			// - Presence of a specific tag (e.g. "public-dns:yes")
			// - Hostname contains a specific suffix (e.g. ".example.com")
			log.Info().
				Str("requiredSuffix", cfg.RequiredSuffix).
				Str("requiredTags", strings.Join(cfg.RequiredTags, ",")).
				Msg("Collecting eligible Defined.net Managed Nebula hosts")

			hosts, err := FilterHosts(cfg.DefinedNet.APIToken, func(h Host) bool {
				// FIXME check valid fqdn

				// Make sure any required suffix is present
				if !strings.HasSuffix(h.Hostname, cfg.RequiredSuffix) {
					return false
				}

				// Make sure all required tags are present
				hostTags := map[string]struct{}{}
				for _, tag := range h.Tags {
					hostTags[tag] = struct{}{}
				}

				for _, tag := range cfg.RequiredTags {
					if _, ok := hostTags[tag]; !ok {
						return false
					}
				}

				return true
			})
			if err != nil {
				return fmt.Errorf("failed to collect eligible hosts: %w", err)
			}

			log.Info().Int("eligibleHosts", len(hosts)).Msgf("Found %d eligible hosts", len(hosts))

			// Create an A record for each host that matches the criteria pointing to
			// the host's IP address. Create a map of valid hostnames as we go.
			hostnames := map[string]struct{}{}
			for _, host := range hosts {
				hostname := host.Hostname
				l := log.Info().Str("initialHostname", hostname)
				if cfg.TrimSuffix {
					hostname = trimSuffix(hostname)
					l = l.Str("trimmedHostname", hostname)
				}
				hostname = strings.ToLower(hostname + "." + cfg.AppendSuffix)
				l.Str("finalHostname", hostname).
					Str("ipAddress", host.IPAddress).
					Msg("Creating Cloudflare DNS record")

				err := CreateRecord(cf, zoneID, hostname, host.IPAddress)
				if err != nil {
					// TODO: Log the error and continue
					return fmt.Errorf("failed to create record: %w", err)
				}

				hostnames[hostname] = struct{}{}
			}

			// For any hosts within the target zone that do not have a corresponding
			// host in Defined Networking, delete the A record
			if cfg.Prune == "all" || cfg.Prune == "network" {
				log.Info().Str("zoneID", zoneID).Str("mode", cfg.Prune).
					Msg("Pruning Cloudflare DNS records")

				err := IterateRecords(cf, zoneID, func(r Record) error {
					if !strings.HasSuffix(r.Name, cfg.AppendSuffix) {
						return nil
					}

					if _, ok := hostnames[r.Name]; !ok {
						// In network mode, only prune A records within the Nebula CIDR
						if cfg.Prune == "network" {
							if r.Type != "A" {
								return nil
							}

							ip, err := netip.ParseAddr(r.Content)
							if err != nil {
								return fmt.Errorf("failed to parse IP address from record content: %w", err)
							}
							if !cidr.Contains(ip) {
								return nil
							}
						}

						log.Info().Str("recordID", r.ID).
							Str("recordName", r.Name).
							Msg("Pruning stale DNS record")

						err := DeleteRecord(cf, zoneID, r.ID)
						if err != nil {
							return fmt.Errorf("failed to delete record: %w", err)
						}
					}

					return nil
				})
				if err != nil {
					return fmt.Errorf("error during host prune iteration: %w", err)
				}
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		return err
	}

	return nil
}

func trimSuffix(s string) string {
	if idx := strings.Index(s, "."); idx != -1 {
		return s[:idx]
	}
	return s
}
