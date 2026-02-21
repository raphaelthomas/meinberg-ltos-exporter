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

// CheckHealth checks if the Meinberg device is reachable
func (c *Client) CheckHealth() (bool, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/status", nil)
	if err != nil {
		return false, err
	}

	// Apply authentication
	if c.authBasicUser != "" && c.authBasicPass != "" {
		req.SetBasicAuth(c.authBasicUser, c.authBasicPass)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Discard response body
	_, _ = io.ReadAll(resp.Body)

	return true, nil
}

// FetchStatus fetches the system status from the LTOS API
func (c *Client) FetchStatus() (map[string]interface{}, error) {
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

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return data, nil
}
