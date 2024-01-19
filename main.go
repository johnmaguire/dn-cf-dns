// dn-cf-dns is a tool for creating DNS records in Cloudflare based on hosts
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
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

func main() {
	if err := mainWithErr(); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

func mainWithErr() error {
	cmd := &cli.Command{
		Name:        "dn-cf-dns",
		Version:     "0.1.0",
		Description: "dn-cf-dns manages DNS records in Cloudflare based on Defined Networking hosts",
		Authors: []any{
			&mail.Address{Name: "John Maguire", Address: "contact@johnmaguire.me"},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Value: "./config.toml", Usage: "path to config file"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			// Read the config file
			cfg, err := LoadConfig(c.String("config"))
			if err != nil {
				return err
			}

			fmt.Printf("%#v\n", cfg)

			// Filter the DN hosts based on the following criteria:
			// - Presence of a specific tag (e.g. "public-dns:yes")
			// - Hostname contains a specific suffix (e.g. ".example.com")
			hosts, err := FilterHosts(cfg.DefinedNet.APIToken, func(h Host) bool {
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
				return err
			}

			// Create an A record for each host that matches the criteria pointing to
			// the host's IP address. Create a map of valid hostnames as we go.
			hostnames := map[string]struct{}{}
			for _, host := range hosts {
				hostname := host.Hostname
				if cfg.TrimSuffix {
					hostname = trimSuffix(hostname)
				}

				err := CreateRecord(cfg.Cloudflare.APIToken, cfg.Cloudflare.ZoneID, hostname, host.IPAddress)
				if err != nil {
					// TODO: Log the error and continue
					return err
				}

				hostnames[hostname] = struct{}{}
			}

			// For any hosts within the target zone that do not have a corresponding
			// host in Defined Networking, delete the A record
			if cfg.PruneRecords {
				err := IterateRecords(cfg.Cloudflare.APIToken, cfg.Cloudflare.ZoneID, func(r Record) error {
					if _, ok := hostnames[r.Name]; !ok {
						return DeleteRecord(cfg.Cloudflare.APIToken, cfg.Cloudflare.ZoneID, r.ID)
					}

					return nil
				})
				if err != nil {
					return err
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
