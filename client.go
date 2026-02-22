// Copyright 2026 Raphael Seebacher
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Client represents a Meinberg LTOS API client
type Client struct {
	baseURL       string
	timeout       time.Duration
	authBasicUser string
	authBasicPass string
	httpClient    *http.Client
}

// parseCPULoad parses the cpuload string and returns the 1, 5, and 15 minute averages
// Example input: "0.48 0.66 0.57 2/99 25157"
func parseCPULoad(cpuloadStr string) (float64, float64, float64, error) {
	parts := strings.Fields(cpuloadStr)
	if len(parts) < 3 {
		return 0, 0, 0, fmt.Errorf("failed to parse CPU load string: %q", cpuloadStr)
	}

	load1, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse 1-minute CPU load: %v", err)
	}
	load5, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse 5-minute CPU load: %v", err)
	}
	load15, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse 15-minute CPU load: %v", err)
	}

	return load1, load5, load15, nil
}

// parseMemory parses the memory string and returns total and free memory in bytes.
// Example input: "228428 kB total memory, 161732 kB free (70 %)"
// Returns an error if the memory string cannot be parsed.
func parseMemory(memoryStr string) (float64, float64, error) {
	// Extract total memory (first number)
	totalRe := regexp.MustCompile(`(\d+)\s+kB\s+total`)
	totalMatches := totalRe.FindStringSubmatch(memoryStr)
	if len(totalMatches) < 2 {
		return 0, 0, fmt.Errorf("failed to parse total memory: %q", memoryStr)
	}
	totalMemoryKB, err := strconv.ParseFloat(totalMatches[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse total memory as float: %v", err)
	}

	// Extract free memory (second number)
	freeRe := regexp.MustCompile(`(\d+)\s+kB\s+free`)
	freeMatches := freeRe.FindStringSubmatch(memoryStr)
	if len(freeMatches) < 2 {
		return 0, 0, fmt.Errorf("failed to parse free memory: %q", memoryStr)
	}
	freeMemoryKB, err := strconv.ParseFloat(freeMatches[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse free memory as float: %v", err)
	}

	// Convert from KB to bytes
	return totalMemoryKB * 1024, freeMemoryKB * 1024, nil
}

// Target returns the target base URL of the Meinberg LTOS API client
func (c *Client) Target() string {
	return c.baseURL
}

// NewClient creates a new Meinberg LTOS API client
func NewClient(baseURL string, timeout time.Duration, authBasicUser, authBasicPass string) *Client {
	return &Client{
		baseURL:       baseURL,
		timeout:       timeout,
		authBasicUser: authBasicUser,
		authBasicPass: authBasicPass,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// FetchStatus fetches the target status from the Meinberg LTOS API
func (c *Client) FetchStatus() (map[string]any, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/status", nil)
	if err != nil {
		return nil, err
	}

	// Apply authentication
	if c.authBasicUser != "" && c.authBasicPass != "" {
		req.SetBasicAuth(c.authBasicUser, c.authBasicPass)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return data, nil
}
