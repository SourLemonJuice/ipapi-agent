package upstream

import (
	"fmt"

	"github.com/SourLemonJuice/ipapi-agent/internal/response"
)

/*
Docs: https://ipapi.co/api/#complete-location
Example: https://ipapi.co/1.1.1.1/json/

	{
	  "ip": "1.1.1.1",
	  "network": "1.1.1.0/24",
	  "version": "IPv4",
	  "city": "Sydney",
	  "region": "New South Wales",
	  "region_code": "NSW",
	  "country": "AU",
	  "country_name": "Australia",
	  "country_code": "AU",
	  "country_code_iso3": "AUS",
	  "country_capital": "Canberra",
	  "country_tld": ".au",
	  "continent_code": "OC",
	  "in_eu": false,
	  "postal": "2000",
	  "latitude": -33.859336,
	  "longitude": 151.203624,
	  "timezone": "Australia/Sydney",
	  "utc_offset": "+1000",
	  "country_calling_code": "+61",
	  "currency": "AUD",
	  "currency_name": "Dollar",
	  "languages": "en-AU",
	  "country_area": 7686850,
	  "country_population": 24992369,
	  "asn": "AS13335",
	  "org": "CLOUDFLARENET"
	}
*/
type ipapiCo struct {
	IP                 string  `json:"ip"`
	City               string  `json:"city"`
	Region             string  `json:"region"`
	RegionCode         string  `json:"region_code"`
	Country            string  `json:"country"`
	CountryCode        string  `json:"country_code"`
	CountryCodeISO3    string  `json:"country_code_iso3"`
	CountryName        string  `json:"country_name"`
	CountryCapital     string  `json:"country_capital"`
	CountryTLD         string  `json:"country_tld"`
	CountryArea        float32 `json:"country_area"`
	CountryPopulation  string  `json:"country_population"`
	ContinentCode      string  `json:"continent_code"`
	InEU               bool    `json:"in_eu"`
	Postal             string  `json:"postal"`
	Latitude           float32 `json:"latitude"`
	Longitude          float32 `json:"longitude"`
	Latlong            string  `json:"latlong"`
	Timezone           string  `json:"timezone"`
	UTCOffset          string  `json:"utc_offset"`
	CountryCallingCode string  `json:"country_calling_code"`
	Currency           string  `json:"currency"`
	CurrencyName       string  `json:"currency_name"`
	Languages          string  `json:"languages"`
	ASN                string  `json:"asn"`
	Org                string  `json:"org"`
	Hostname           string  `json:"hostname"`
}

func (data *ipapiCo) DoRequest(addr string) error {
	err := getJSON(fmt.Sprintf("https://ipapi.co/%v/json/", addr), data)
	if err != nil {
		return err
	}

	return nil
}

func (data *ipapiCo) Fill(resp *response.Query) error {
	var err error

	resp.DataSource = "ipapi.co"
	resp.Country = data.CountryName
	resp.CountryCode = data.CountryCode
	resp.Region = data.Region
	resp.Timezone = data.Timezone

	// UTCOffset
	resp.UTCOffset, err = timezoneToUTCOffset(data.Timezone)
	if err != nil {
		return fmt.Errorf("can not convert UTC offset: %w", err)
	}

	resp.ISP = data.Org // no ISP data available
	resp.Org = data.Org
	resp.ASN = data.ASN

	return nil
}
