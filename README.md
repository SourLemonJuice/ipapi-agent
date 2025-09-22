# IPAPI-agent

Reverse proxy public IP data API info.

## Configuration

IPAPI-agent use TOML as the config file format. You can specific the config file path via command line flag:

```shell
ipapi-agent --config ./config.toml
```

Typically, your config may look like this:

```toml
listen = "::"
listen_port = 8080
resolve_domain = true
# trusted_proxies = ["127.0.0.1", "::1"]
```

## Config Top-Level section

### listen

`string`\
Let server listen to which IP address?

Default: `listen = "::"`

### listen_port

`uint16`\
Let server listen to which port(TCP/HTTP1.1)?

Default: `listen_port = 8080`

### resolve_domain

`bool`\
Controls whether domain name resolution is permitted.

Default: `resolve_domain = true`

### trusted_proxies

`string list`\
Controls which IP addresses or CIDRs can use `X-Real-IP`, this should be a reverse proxy.

Default: `trusted_proxies = ["127.0.0.1", "::1"]`

## Config [dev] section

Unstable entries, must not use these configs in production.

### dev.debug

`bool`\
GIN debug mode

Default: `debug = false`

### dev.log

`bool`\
GIN log, other logs are not affected.

Default: `log = false`

### dev.tld_blocklist

`string list`\
Extend the TLD blocklist used to resolve the domain. You may want to block `.lan` TLD at here, which it supported by some home routers DHCP server but not a part of any standard.

Default: none
