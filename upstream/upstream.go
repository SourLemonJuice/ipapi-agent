package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/SourLemonJuice/ipapi-agent/response"
)

type API interface {
	// Request the upstream API and return a Query structure without "status" and "message" filed.
	// If failure, response InternalServerError.
	Fetch(ctx context.Context, addr string) (response.Query, error)
}

func fetchJSON(ctx context.Context, url string, data API) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request error: %v", err)
	}

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
	tz, err := time.LoadLocation(tzStr)
	if err != nil {
		return 0, fmt.Errorf("can not load timezone: %w", err)
	}

	_, offset_sec := time.Now().In(tz).Zone()
	return offset_sec / 60, nil
}
