# IPAPI-Agent API Reference

## GET `/query/<IP addr or domain>`

Response JSON:

|Name|Description|Example|Type|
|--|--|--|--|
|status|`success` or `failure`|`"success"`|string|
|message|User-friendly message, **ONLY exists** when failure state. unsure content|`"Data source error"`|string|
|dataSource|One of upstream data providers: `ip-api`|`"ip-api"`|string|
|country|Country common name|`"United Kingdom"`|string|
|countryCode|ISO 3166 Country two-letters code|`"GB"`|string|
|region|Region name|`"England"`|string|
|timezone|Timezone information|`"Europe/London"`|string|
|utcOffset|Kind like [getTimezoneOffset()](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date/getTimezoneOffset) but result "UTC+" with positive "UTC-" with negative. Unit is minutes|`480`(UTC+8)|int|
|isp|Internet service provider(ISP) name|`"Sky UK Limited"`|string|
|org|Organization name|`"Sky Broadband"`|string|
|asn|Autonomous System Number|`"AS5607"`|string|
