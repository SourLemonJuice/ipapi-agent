package datasource

import (
	"fmt"

	"github.com/SourLemonJuice/ipapi-agent/resps"
)

type IpapiCo struct {
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

func (data *IpapiCo) DoRequest(addr string) error {
	err := getJSON(fmt.Sprintf("https://ipapi.co/%v/json/", addr), data)
	if err != nil {
		return err
	}

	return nil
}

func (data *IpapiCo) Fill(resp *resps.Query) error {
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
