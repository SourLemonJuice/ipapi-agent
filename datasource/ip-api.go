package datasource

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/SourLemonJuice/ipapi-agent/respstruct"
)

type IpapiCom struct {
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	RegionName  string `json:"regionName"`
	Timezone    string `json:"timezone"`
	ISP         string `json:"isp"`
	Org         string `json:"org"`
	AS          string `json:"as"`
}

// If failure, response OK with JSON message
func (data *IpapiCom) DoRequest(addr string) error {
	var err error

	resp, err := http.Get("http://ip-api.com/json/" + addr + "?fields=53003")
	if err != nil {
		return fmt.Errorf("HTTP client error: %w", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return fmt.Errorf("JSON parsing error: %w", err)
	}

	switch data.Status {
	case "success":
	case "fail":
		return fmt.Errorf("data source response error: %v", data.Message)
	default:
		// for security considered, the undefined status shouldn't be returned
		return errors.New("unknown data source response status")
	}

	return nil
}

// If failure, response InternalServerError
func (data *IpapiCom) Fill(resp *respstruct.Query) error {
	var err error

	resp.DataSource = "ip-api.com"
	resp.Country = data.Country
	resp.CountryCode = data.CountryCode
	resp.Region = data.RegionName
	resp.Timezone = data.Timezone

	// UTCOffset
	resp.UTCOffset, err = data.getUTCOffset()
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

func (data *IpapiCom) getUTCOffset() (int, error) {
	var err error

	tz, err := time.LoadLocation(data.Timezone)
	if err != nil {
		return 0, fmt.Errorf("can not load API returned timezone: %w", err)
	}

	_, offset_sec := time.Now().In(tz).Zone()
	return offset_sec / 60, nil
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
