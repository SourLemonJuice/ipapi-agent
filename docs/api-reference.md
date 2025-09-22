# IPAPI-agent API Reference

## GET `/`

Respond with plain text in a human-friendly format.\
The content of the response is uncertain, you should not use this endpoint in your script.

The content of the response is almost a variant of [`/query/<IP addr or domain>`](#get-queryip-addr-or-domain), but queries the client IP address.\
Its limitations also apply.

This endpoint shares the query cache with [`/query/<IP addr or domain>`](#get-queryip-addr-or-domain), but can not be disabled.

Example:

```text
● 1.1.1.1 | ip-api.com
 Location: Queensland, Australia (AU)
 Timezone: Australia/Brisbane UTC+1000
      ISP: Cloudflare, Inc
      Org: APNIC and Cloudflare DNS Resolver project
      ASN: AS13335
```

Try it with curl:

```shell
curl https://ipapi.example.com
# or request a fake client IP via X-Real-IP, as long as you are in the trusted_proxies list
curl -H 'X-Real-IP: 1.1.1.1' localhost:8080
```

## GET `/query/<IP addr or domain>`

Response JSON:

|Name|Description|Example|Type|
|--|--|--|--|
|status|`success` or `failure`|`"success"`|string|
|message|User-friendly message, **ONLY exists** when failure state. Uncertain content|`"Data source error"`|string|
|dataSource|One of upstream data providers: `ip-api.com`|`"ip-api.com"`|string|
|country|Country common name|`"United Kingdom"`|string|
|countryCode|ISO 3166 Country two-letters code|`"GB"`|string|
|region|Region name|`"England"`|string|
|timezone|Timezone information|`"Europe/London"`|string|
|utcOffset|Kind like [getTimezoneOffset()](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date/getTimezoneOffset) but result "UTC+" with positive "UTC-" with negative. Unit is minutes|`480`(UTC+8)|int|
|isp|Internet service provider(ISP) name|`"Sky UK Limited"`|string|
|org|Organization name|`"Sky Broadband"`|string|
|asn|Autonomous System Number|`"AS5607"`|string|

Query strings:

|Name|Description|Example|Value Range|
|--|--|--|--|
|cache|Force control whether the server uses its cache|`cache=false`|`true` or `false`|

> Note: Request a loopback, private, unspecified(0.0.0.0/::), or any non-global unicast address will return an error(status `failure`).\
> Even though, many reserved addresses/CIDRs are still not filtered.

If you are querying a reserved domain, it will also return an error. You can extend this list in config file.\
Current list: `".alt", ".arpa", ".invalid", ".local", ".localhost", ".onion", ".test", ".internal"`\
Source: [Special-use domain name - Wikipedia](https://en.wikipedia.org/wiki/Special-use_domain_name)

Consider that some DNS servers will respond with a geolocation-related IP address to reduce CDN's loading time.\
If you still feel resolving a domain is dangerous, you can set `resolve_domain = false` in config file to protect your server location securely.

## GET `/query`

Same as above, but request your client IP address.
