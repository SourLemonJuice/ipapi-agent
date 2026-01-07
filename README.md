# IPAPI-agent

Reverse proxy public IP data API.\
View the IP country/region and ASN in a JSON or plain text format.

## Usage/API Reference

Try our human interface via curl:

```sh
curl ip.charchar.dev
curl ip.charchar.dev/1.1.1.1
```

Or check out the full REST api reference(include details of the human interface): [docs/api-reference.md](docs/api-reference.md)

## Configuration

IPAPI-agent use TOML as the config file format. It'll automatically find `ipapi.toml` file in the same directory, or you can specific the config file path via the command line flag:

```sh
ipapi-agent --config ./ipapi.toml
```

For more examples, see: [ipapi.toml.example](ipapi.toml.example)

## Deployment

The GitHub CI will automatically build and push the amd64/arm64/riscv64 containers to the registries. You can pull those containers from:

| Registry | URL |
| --- | --- |
| Docker Hub | [sourlemonjuice/ipapi-agent](https://hub.docker.com/r/sourlemonjuice/ipapi-agent) |
| GitHub Container Registry | [ghcr.io/sourlemonjuice/ipapi-agent](https://github.com/SourLemonJuice/ipapi-agent/pkgs/container/ipapi-agent) |

The example Docker Compose file:

```yaml
name: ipapi
services:
  ipapi-agent:
    image: sourlemonjuice/ipapi-agent:latest
    restart: unless-stopped
    ports:
      - 8080:8080
    volumes:
      - ./ipapi.toml:/ipapi.toml:ro
```

Or this command:

```sh
docker run --rm -it -p 8080:8080 -v ./ipapi.toml:/ipapi.toml:ro sourlemonjuice/ipapi-agent:latest
```

> [!NOTE]
> This container is CGO disabled and based on Alpine, don't worry about container size :)\
> Also, after v0.5.0, you can use a sematic tag like `0.5` to reference the latest version of v0.5.* or set to full version `0.5.0`. This replaced old git tag based naming: `v0.4.1`.

## Config Top-Level section

### listen `string`

An IPv4 or IPv6 address to use for server listening.\
Default: `listen = "::"`

### port `uint16`

Server listening port.\
Default: `port = 8080`

> [!NOTE]
> Changing the port in container will break the health check. Leave this unset/default.

### trusted_proxies `string list`

Controls which IP addresses or CIDRs can use `X-Forwarded-For` or `X-Real-IP`, this should be a reverse proxy.\
Default: `trusted_proxies = ["127.0.0.1", "::1"]`

## Config [upstream] section

### upstream.mode `string`

Upstream selection mode. Available values: `single`, `random`, `rotated`.

*single*: only use your only one upstream or the first one in the list.

*random*: randomly choice upstream from the upstream list, per-request applied.

*rotate*: keep rotating the upstream from the upstream list. The interval time can be set with `rotate_interval` below.

> [!NOTE]
> Whatever the mode of selection, the cache system will not be affected at all.\
> For example, if the cache time-to-live is 6 hours, during these 6 hours the responses all come from one upstream in a cache pool.

Default: `mode = "single"`

### upstream.pool `string/string list`

Set one or more upstreams for further selection. Available codenames:

- `ip-api.com`: very normal option and feel reliable, preferred.\
  Docs: <https://ip-api.com/docs/api:json>
- `ipinfo-free`: also preferred, but they say this is a *legacy/free* API :)\
  Docs: <https://ipinfo.io/missingauth>
- `ipapi.co`: free rate limit to 100 requests per month... Added just for fun.\
  Docs: <https://ipapi.co/api/#complete-location>

Default: `pool = "ipinfo-free"`\
You can also: `pool = ["ip-api.com", "ipinfo-free"]`

### upstream.rotate_interval `string`

Upstream rotation interval used in `rotate` mode. Parse with go's [time.ParseDuration()](https://pkg.go.dev/time#ParseDuration).

Default: `rotate_interval = "1h"`\
You can also: `rotate_interval = "72h99m23s"`

## Config [domain] section

### domain.enabled `bool`

Controls whether domain name resolution is permitted.\
Default: `enabled = true`

### domain.block_suffix `string list`

Extend the domain public suffix(not only TLD) blocklist used when resolving the domain. You may want to block `lan` TLD here, which it supported by some home routers DHCP server but standard.

Built-in list is: `"alt", "arpa", "invalid", "local", "localhost", "onion", "test", "internal"`\
You can also append it: `block_suffix = ["lan"]`

## Config [dev] section

> [!WARNING]
> These entries can not be used in production. Development purpose only.

### dev.debug `bool`

Turn on debug information output(GIN and others).\
Default: `debug = false`

### dev.log `bool`

GIN log, other logs are not affected.\
Default: `log = false`

## License

This software published under Apache-2.0 license.\
Copyright 2025-2026 酸柠檬猹/SourLemonJuice
