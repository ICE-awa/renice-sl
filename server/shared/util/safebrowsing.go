package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const apiEndpoint = "https://safebrowsing.googleapis.com/v4/threatMatches:find"

type SafeBrowsingClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewSafeBrowsingClient(apiKey string) *SafeBrowsingClient {
	return &SafeBrowsingClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *SafeBrowsingClient) IsEnabled() bool {
	return c.apiKey != ""
}

func (c *SafeBrowsingClient) IsSafe(ctx context.Context, rawURL string) (bool, error) {
	reqBody := map[string]any{
		"client": map[string]string{
			"clientId":      "renice-sl",
			"clientVersion": "dev",
		},
		"threatInfo": map[string]any{
			"threatTypes":      []string{"MALWARE", "SOCIAL_ENGINEERING", "UNWANTED_SOFTWARE", "MALICIOUS_BINARY"},
			"platformTypes":    []string{"ANY_PLATFORM"},
			"threatEntryTypes": []string{"URL"},
			"threatEntries": []map[string]string{
				{"url": rawURL},
			},
		},
	}

	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		apiEndpoint+"?key="+c.apiKey, bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("safe browsing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("safe browsing API status: %d", resp.StatusCode)
	}

	var result struct {
		Matches []struct {
			ThreatType string `json:"threatType"`
		} `json:"matches"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("decode response: %w", err)
	}

	return len(result.Matches) == 0, nil
}
