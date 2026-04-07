package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const healthClientTimeout = 10 * time.Second

func checkHealth(rawURL string) error {
	u := strings.TrimSpace(rawURL)
	if u == "" {
		return fmt.Errorf("url is empty")
	}
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		return fmt.Errorf("url must start with http:// or https://")
	}
	client := &http.Client{Timeout: healthClientTimeout}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}
