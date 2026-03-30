# nebula-dns

A script to create and delete [Cloudflare](https://cloudflare.com) DNS records pointing to [Defined Networking](https://defined.net) Managed Nebula nodes.

## Development

To build the tool, run `make bin`. This will create a `nebula-dns` binary in the root directory.

To run tests, run `make test`.

## Usage

Copy `examples/config.toml` to `config.toml` and fill in the values, then run `./nebula-dns` to create the DNS records.
