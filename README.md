# IPAPI-agent

Reverse proxy public IP data API.\
View the IP country/region and ASN in a JSON or plain text format.

## Usage/API Reference

Try our human interface via curl:

```sh
curl https://ip.charchar.dev
```

Or check out the full REST api reference(include details of the human interface): [docs/api-reference.md](docs/api-reference.md)

## Configuration

IPAPI-agent use TOML as the config file format. It'll automatically find `ipapi.toml` file in the same directory, or you can specific the config file path via the command line flag:

```sh
ipapi-agent --config ./ipapi.toml
```

Typically, your config may look like this:

```toml
listen = "::"
port = 8080
# use this list in docker container: ["172.16.0.0/12"]
trusted_proxies = ["127.0.0.1", "::1"]
```

These are the default values, you can also run it without config file.\
For more examples, see: [ipapi.toml.example](ipapi.toml.example)

## Deployment

The GitHub CI will automatically build and push the amd64/arm64/riscv64 containers to the registries. You can pull those containers from:

| Registry | URL |
| --- | --- |
|Docker Hub|[sourlemonjuice/ipapi-agent](https://hub.docker.com/r/sourlemonjuice/ipapi-agent)|
|GitHub Container Registry|[ghcr.io/sourlemonjuice/ipapi-agent](https://github.com/SourLemonJuice/ipapi-agent/pkgs/container/ipapi-agent)|

The Docker Compose file can reference this simple example: [compose.yaml](compose.yaml).\
After v0.5.0, you can use a sematic tag like `0.5` to reference the latest version of v0.5.* or set to full version `0.5.0`. This replaced old git tag based naming: `v0.4.1`

> [!NOTE]
> This container is CGO disabled and based on scratch. The package size is very small, but you won't be able to use many utils inside the container for debugging.

## Config Top-Level section

### listen

`string` An IPv4 or IPv6 address to use for server listening.\
Default: `listen = "::"`

### port

`uint16` Server listening port.\
Default: `port = 8080`

### trusted_proxies

`string list` Controls which IP addresses or CIDRs can use `X-Real-IP`, this should be a reverse proxy.\
Default: `trusted_proxies = ["127.0.0.1", "::1"]`

## Config [upstream] section

### upstream.mode

`string` Upstream selection mode. Available values: `single`, `random`, `rotated`.

*single*: only use your only one upstream or the first one in the list.

*random*: randomly choice upstream from the upstream list, per-request applied.

*rotated* (broken): keep rotating the upstream from the upstream list. The interval time can be set with `rotated_interval` below.\
Note, interval time is kind of like a minimal gap of the upstream renewal. The timer will only restarted after a user request.

> [!NOTE]
> Whatever the mode of selection, the cache system will not be affected at all.\
> For example, if the cache time-to-live is 6 hours, during these 6 hours the responses all come from one upstream in a cache pool.

Default: `mode = "single"`

### upstream.pool

`string/string list` Set one or more upstreams for further selection. Available codenames:

- `ip-api.com`: very normal option and feel reliable, preferred.\
  Docs: <https://ip-api.com/docs/api:json>
- `ipinfo-free`: also preferred, but they say this is a *legacy/free* API :)\
  Docs: <https://ipinfo.io/missingauth>
- `ipapi.co`: free rate limit to 100 requests per month... Added just for fun.\
  Docs: <https://ipapi.co/api/#complete-location>

Default: `pool = "ip-api.com"`\
You can also: `pool = ["ip-api.com", "ipinfo-free"]`

### upstream.rotated_interval

`string` Upstream rotation interval used in `rotated` mode. Parse with go's [time.ParseDuration()](https://pkg.go.dev/time#ParseDuration).

Default: `rotated_interval = "24h"`\
You can also: `rotated_interval = "72h99m23s"`

## Config [resolve] section

### resolve.domain

`bool` Controls whether domain name resolution is permitted.\
Default: `domain = true`

### resolve.block_tld

`string list` Extend the TLD blocklist used to resolve the domain. You may want to block `.lan` TLD at here, which it supported by some home routers DHCP server but not a part of any standard.

Default: none\
You can also: `block_tld = [".lan"]`

## Config [dev] section

> [!WARNING]
> UNSTABLE entries, must not use these configs in production. Only use for development purpose.

### dev.debug

`bool` Turn on debug information output(GIN and others).\
Default: `debug = false`

### dev.log

`bool` GIN log, other logs are not affected.\
Default: `log = false`

## License

This software published under Apache-2.0 license.\
Copyright 2025 酸柠檬猹/SourLemonJuice
