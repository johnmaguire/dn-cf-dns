# nebula-dns

A script to create and delete [Cloudflare](https://cloudflare.com) DNS records pointing to [Defined Networking](https://defined.net) Managed Nebula nodes.

## Development

To build the tool, run `make bin`. This will create a `nebula-dns` binary in the root directory.

To run tests, run `make test`.

## Usage

### Config file

Copy `examples/config.toml` to `config.toml` and fill in the values, then run `./nebula-dns --config config.toml`.

### Environment variables

All configuration can be provided via environment variables, making the config file optional. Environment variables take precedence over values in the config file.

| Variable | Description |
|---|---|
| `NEBULA_DNS_CF_API_TOKEN` | Cloudflare API token (`Zone:DNS:Edit` permission) |
| `NEBULA_DNS_CF_ZONE_NAME` | Cloudflare DNS zone (e.g. `example.com`) |
| `NEBULA_DNS_DN_API_TOKEN` | Defined Networking API token (`hosts:list` permission) |
| `NEBULA_DNS_DN_NETWORK_ID` | Defined Networking network ID |
| `NEBULA_DNS_REQUIRED_TAGS` | Comma-separated list of required host tags |
| `NEBULA_DNS_REQUIRED_SUFFIX` | Only register hosts with this hostname suffix |
| `NEBULA_DNS_TRIM_SUFFIX` | Trim domain from DN hostname (`true`/`false`) |
| `NEBULA_DNS_APPEND_SUFFIX` | Suffix to append to hostname (defaults to zone name) |
| `NEBULA_DNS_PRUNE_RECORDS` | Delete stale DNS records (`true`/`false`) |
| `NEBULA_DNS_PRUNE_NETWORK_RECORDS_ONLY` | Only prune records within the network CIDR (`true`/`false`) |
