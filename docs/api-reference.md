# IPAPI-agent API Reference

## GET `/`

Respond with plain text in a human-friendly format. The request IP address is your client address.\
The content of the response is uncertain, you should not use this endpoint in your automation script.

The content of the response is almost a variant of [`/query/<IP addr or domain>`](#get-queryip-addr-or-domain), but queries the client IP address.\
Its limitations also apply.

This endpoint shares the query cache with [`/query/<IP addr or domain>`](#get-queryip-addr-or-domain), but can not be disabled.

For example:

```text
● 1.1.1.1 (Anycast) - IPinfo Free
  Location: Hong Kong, Hong Kong (Special Administrative Region of China) (HK)
  Timezone: Asia/Hong_Kong UTC+0800
       Org: Cloudflare, Inc.
       ISP: Cloudflare, Inc.
       ASN: AS13335
```

Or this:

```text
× FAILURE
IP address/domain is in invalid range
```

Try it with curl:

```shell
curl ip.charchar.dev
# or request a fake client IP via X-Forwarded-For, as long as you are in the trusted_proxies list
curl -H 'X-Forwarded-For: 1.1.1.1' localhost:8080
```

After v0.2.0, if the user agent of the client is `curl`, some ANSI color codes will be added. \awa/\
I copied those ideas from systemd, haha.

## GET `/<IP addr or domain>`

Same as `/`, but responded with your given IP address.

## GET `/query/<IP addr or domain>`

Response JSON:

|Name|Description|Example|Type|
|--|--|--|--|
|status|`success` or `failure`|`"success"`|string|
|message|User-friendly message, **ONLY exists** when failure state. Uncertain content|`"Data source error"`|string|
|dataSource|One of upstream data providers: `ipinfo-free`, `ip-api.com`, `ipapi.co`|`"ipinfo-free"`|string|
|country|Country common name|`"United Kingdom"`|string|
|countryCode|ISO 3166 Country two-letters code|`"GB"`|string|
|region|Region name|`"England"`|string|
|timezone|Timezone information|`"Europe/London"`|string|
|utcOffset|Kind like [getTimezoneOffset()](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date/getTimezoneOffset) but result "UTC+" with positive "UTC-" with negative. Unit is minutes|`480`(UTC+8)|int|
|org|Organization name|`"Sky Broadband"`|string|
|isp|Internet service provider(ISP) name|`"Sky UK Limited"`|string|
|asn|Autonomous System Number|`"AS5607"`|string|
|anycast|Anycast info, only available when using `ipinfo-free`|`true`|bool|

Query strings:

|Name|Description|Example|Value Range|
|--|--|--|--|
|cache|Force control whether the server uses its cache|`cache=false`|`true` or `false`|

> Note: Request a loopback, private, unspecified(0.0.0.0/::), or any non-global unicast address will return an error(status `failure`).\
> Even though, many reserved addresses/CIDRs are still not filtered.

If you are querying a reserved domain, it will also return an error. You can extend this list in the config file(see `[domain]` section).\
Source: [Special-use domain name - Wikipedia](https://en.wikipedia.org/wiki/Special-use_domain_name)

Consider that some DNS servers will respond with a geolocation-related IP address to reduce CDN's loading time.\
If you still feel resolving a domain is dangerous, you can set `resolve.domain = false` in config file to protect your server location securely.

## GET `/query`

Same as above, but responded with your client IP address.

## GET `/generate_204`

Health check, always return HTTP 204 NO CONTENT.
