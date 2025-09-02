# IPAPI-Agent API Reference

## General Response Type

Generally, both `success` or `failure` status will response as *HTTP 200 OK*.\
But in some case, the server also will response *400 Bad Request* or *500 Internal Server Error* without JSON body.

## GET `/query`

Same as below, but request your client IP address.

## GET `/query/<IP addr or domain>`

Response JSON:

|Name|Description|Example|Type|
|--|--|--|--|
|status|`success` or `failure`|`"success"`|string|
|message|User-friendly message, **ONLY exists** when failure state. unsure content|`"Data source error"`|string|
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

> Note: Requesting a loopback, private, unspecified(0.0.0.0/::), or any non-global unicast address will return an error(status `failure`).
