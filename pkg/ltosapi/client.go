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

package ltosapi

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/ltosapi/models"
)

// Client represents a Meinberg LTOS API client
type Client struct {
	logger        *slog.Logger
	baseURL       string
	authBasicUser string
	authBasicPass string
	httpClient    *http.Client
}

// Target returns the target base URL of the Meinberg LTOS API client
func (c *Client) Target() string {
	return c.baseURL
}

// NewClient creates a new Meinberg LTOS API client
func NewClient(baseURL string, timeout time.Duration, authBasicUser, authBasicPass string, ignoreSSLVerify bool, logger *slog.Logger) *Client {
	return &Client{
		logger:        logger,
		baseURL:       baseURL,
		authBasicUser: authBasicUser,
		authBasicPass: authBasicPass,
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreSSLVerify},
			},
		},
	}
}

// FetchStatus fetches the target status from the Meinberg LTOS API
func (c *Client) FetchStatus() (*models.StatusResponse, error) {
	c.logger.Debug("Fetching status from Meinberg LTOS device API", "url", c.baseURL+"/api/status")

	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/api/status", nil)
	if err != nil {
		return nil, err
	}

	if c.authBasicUser != "" && c.authBasicPass != "" {
		req.SetBasicAuth(c.authBasicUser, c.authBasicPass)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("Unexpected status code from Meinberg LTOS device API", "url", c.baseURL+"/api/status", "status_code", resp.StatusCode)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data models.StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status response: %w", err)
	}

	c.logger.Debug("Successfully fetched status from Meinberg LTOS device API", "url", c.baseURL+"/api/status")
	return &data, nil
}
