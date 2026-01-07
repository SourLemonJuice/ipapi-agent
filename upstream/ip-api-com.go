package upstream

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/SourLemonJuice/ipapi-agent/response"
)

/*
Docs: https://ip-api.com/docs/api:json
Example: http://ip-api.com/json/1.1.1.1?fields=53003

	{
	  "status": "success",
	  "country": "Australia",
	  "countryCode": "AU",
	  "regionName": "Queensland",
	  "timezone": "Australia/Brisbane",
	  "isp": "Cloudflare, Inc",
	  "org": "APNIC and Cloudflare DNS Resolver project",
	  "as": "AS13335 Cloudflare, Inc."
	}
*/
type ipApiCom struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	RegionName  string `json:"regionName"`
	Timezone    string `json:"timezone"`
	ISP         string `json:"isp"`
	Org         string `json:"org"`
	AS          string `json:"as"`
}

func (data *ipApiCom) Fetch(ctx context.Context, addr string) (resp response.Query, err error) {
	err = fetchJSON(ctx, fmt.Sprintf("http://ip-api.com/json/%v?fields=53003", addr), data)
	if err != nil {
		return resp, err
	}

	switch data.Status {
	case "success":
	case "fail":
		return resp, fmt.Errorf("response error: %v", data.Message)
	default:
		// for security considered, the undefined status shouldn't be returned
		return resp, errors.New("unknown response status")
	}

	resp.DataSource = "ip-api.com"
	resp.Country = data.Country
	resp.CountryCode = data.CountryCode
	resp.Region = data.RegionName
	resp.Timezone = data.Timezone

	resp.UTCOffset, err = timezoneToUTCOffset(data.Timezone)
	if err != nil {
		return resp, fmt.Errorf("can not convert UTC offset: %w", err)
	}

	resp.Org = data.Org
	resp.ISP = data.ISP
	resp.ASN, err = data.getASN()
	if err != nil {
		return resp, fmt.Errorf("can not convert ASN: %w", err)
	}

	return resp, nil
}

func (data *ipApiCom) getASN() (string, error) {
	if !strings.HasPrefix(data.AS, "AS") {
		return "", errors.New("wrong AS format")
	}

	// spilt AS to get ASN:
	// "AS13335 Cloudflare, Inc."
	before, _, found := strings.Cut(data.AS, " ")
	if !found {
		return "", errors.New("wrong AS format")
	}

	return before, nil
}
