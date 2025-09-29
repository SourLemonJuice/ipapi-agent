package datasource

import (
	"errors"
	"fmt"
	"strings"

	"github.com/SourLemonJuice/ipapi-agent/resps"
)

type IpapiCom struct {
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

func (data *IpapiCom) DoRequest(addr string) error {
	err := getJSON(fmt.Sprintf("http://ip-api.com/json/%v?fields=53003", addr), data)
	if err != nil {
		return err
	}

	switch data.Status {
	case "success":
	case "fail":
		return fmt.Errorf("response error: %v", data.Message)
	default:
		// for security considered, the undefined status shouldn't be returned
		return errors.New("unknown response status")
	}

	return nil
}

func (data *IpapiCom) Fill(resp *resps.Query) error {
	var err error

	resp.DataSource = "ip-api.com"
	resp.Country = data.Country
	resp.CountryCode = data.CountryCode
	resp.Region = data.RegionName
	resp.Timezone = data.Timezone

	// UTCOffset
	resp.UTCOffset, err = timezoneToUTCOffset(data.Timezone)
	if err != nil {
		return fmt.Errorf("can not convert UTC offset: %w", err)
	}

	// ISP
	resp.ISP = data.ISP
	resp.Org = data.Org
	resp.ASN, err = data.getASN()
	if err != nil {
		return fmt.Errorf("can not convert ASN: %w", err)
	}

	return nil
}

func (data *IpapiCom) getASN() (string, error) {
	if !strings.HasPrefix(data.AS, "AS") {
		return "", errors.New("wrong AS format")
	}

	before, _, found := strings.Cut(data.AS, " ")
	if !found {
		return "", errors.New("wrong AS format")
	}

	return before, nil
}
