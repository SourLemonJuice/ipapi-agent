# IPAPI-agent

Reverse proxy public IP data API.\
View the IP country/region and ASN in a JSON or plain text format.

## Configuration

IPAPI-agent use TOML as the config file format. You can specific the config file path via command line flag:

```shell
ipapi-agent --config ./config.toml
```

Typically, your config may look like this:

```toml
listen = "::"
port = 8080
trusted_proxies = ["127.0.0.1", "::1"]
```

These are the default values, you can also run it without config file.\
For more examples, see: [config.toml.example](config.toml.example)

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

## Config [resolve] section

### resolve.domain

`bool` Controls whether domain name resolution is permitted.\
Default: `domain = true`

### resolve.tld_blocklist

`string list` Extend the TLD blocklist used to resolve the domain. You may want to block `.lan` TLD at here, which it supported by some home routers DHCP server but not a part of any standard.\
Default: none

## Config [dev] section

> [!WARNING]
> UNSTABLE entries, must not use these configs in production. Only use for development purpose.

### dev.debug

`bool` GIN debug mode.\
Default: `debug = false`

### dev.log

`bool` GIN log, other logs are not affected.\
Default: `log = false`
