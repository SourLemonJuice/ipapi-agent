package upstream

import (
	"errors"
	"fmt"
	"strings"

	"github.com/SourLemonJuice/ipapi-agent/response"
	"github.com/biter777/countries"
)

/*
Docs: https://ipinfo.io/missingauth
Example: https://ipinfo.io/1.1.1.1/json

	{
		"ip": "1.1.1.1",
		"hostname": "one.one.one.one",
		"city": "Brisbane",
		"region": "Queensland",
		"country": "AU",
		"loc": "-27.4816,153.0175",
		"org": "AS13335 Cloudflare, Inc.",
		"postal": "4101",
		"timezone": "Australia/Brisbane",
		"readme": "https://ipinfo.io/missingauth",
		"anycast": true
	}
*/
type ipinfoFree struct {
	Region   string `json:"region"`
	Country  string `json:"country"`
	Org      string `json:"org"`
	Timezone string `json:"timezone"`
	Anycast  bool   `json:"anycast"`
}

func (data *ipinfoFree) Request(addr string) error {
	err := getJSON(fmt.Sprintf("https://ipinfo.io/%v/json", addr), data)
	if err != nil {
		return err
	}

	return nil
}

func (data *ipinfoFree) Fill(resp *response.Query) error {
	var err error

	resp.DataSource = "IPinfo Free"
	resp.CountryCode = data.Country
	resp.Region = data.Region

	country := countries.ByName(data.Country)
	resp.Country = country.Info().Name

	resp.Timezone = data.Timezone
	resp.UTCOffset, err = timezoneToUTCOffset(data.Timezone)
	if err != nil {
		return fmt.Errorf("can not convert UTC offset: %w", err)
	}

	// split the first space, first part is ASN, second is Org and ISP:
	// "AS13335 Cloudflare, Inc."
	before, after, found := strings.Cut(data.Org, " ")
	if !found {
		return errors.New("wrong organization format of IPinfo Free")
	}
	resp.ASN = before
	resp.Org = after
	resp.ISP = resp.Org
	resp.Anycast = data.Anycast

	return nil
}

func (data *ipinfoFree) Fetch(addr string) (resp response.Query, err error) {
	err = getJSON(fmt.Sprintf("https://ipinfo.io/%v/json", addr), data)
	if err != nil {
		return resp, err
	}

	resp.DataSource = "IPinfo Free"
	resp.CountryCode = data.Country
	resp.Region = data.Region

	country := countries.ByName(data.Country)
	resp.Country = country.Info().Name

	resp.Timezone = data.Timezone
	resp.UTCOffset, err = timezoneToUTCOffset(data.Timezone)
	if err != nil {
		return resp, fmt.Errorf("can not convert UTC offset: %w", err)
	}

	// split the first space, first part is ASN, second is Org and ISP:
	// "AS13335 Cloudflare, Inc."
	before, after, found := strings.Cut(data.Org, " ")
	if !found {
		return resp, errors.New("wrong organization format of IPinfo Free")
	}
	resp.ASN = before
	resp.Org = after
	resp.ISP = resp.Org
	resp.Anycast = data.Anycast

	return resp, nil
}
