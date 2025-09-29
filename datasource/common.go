package datasource

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/SourLemonJuice/ipapi-agent/resps"
)

type Interface interface {
	// Prepare data for the next use.
	// If failure, response InternalServerError with JSON message.
	DoRequest(addr string) error
	// Fill data into the given struct. This won't fill "status" and "message".
	// If failure, response InternalServerError.
	Fill(resp *resps.Query) error
}

func getJSON(url string, data any) error {
	var err error

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("response is not 200 OK: %v", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		return fmt.Errorf("JSON parse error: %w", err)
	}

	return nil
}

func timezoneToUTCOffset(tzStr string) (int, error) {
	var err error

	tz, err := time.LoadLocation(tzStr)
	if err != nil {
		return 0, fmt.Errorf("can not load timezone: %w", err)
	}

	_, offset_sec := time.Now().In(tz).Zone()
	return offset_sec / 60, nil
}
