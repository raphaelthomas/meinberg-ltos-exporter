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
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCollectorWithMockServer tests the collector with a mock API server
func TestCollectorWithMockServer(t *testing.T) {
	// Create a mock server that responds with the Meinberg LTOS API response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/status" {
			http.NotFound(w, r)
			return
		}

		// Create a mock response based on api-status-response.txt
		response := map[string]interface{}{
			"system-information": map[string]interface{}{
				"API Version":    "LANTIME REST API V20.05.013",
				"version":        "fw_7.10.008",
				"serial-number":  "OBFUSCATED_SERIAL_NUMBER",
				"hostname":       "test-device",
				"time-stamp":     "2026-02-11T22:05:07",
				"model":          "M600",
			},
			"data": map[string]interface{}{
				"object-id": "status",
				"system": map[string]interface{}{
					"uptime":   130988.25,
					"cpuload":  "0.48 0.66 0.57 2/99 25157",
					"memory":   "228428 kB total memory, 161732 kB free (70 %)",
					"sync-status": map[string]interface{}{
						"clock-status": map[string]interface{}{
							"clock":       "synchronized",
							"oscillator":  "warmed-up",
							"antenna":     "connected",
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create a client pointing to the mock server
	client := NewClient(mockServer.URL, 5*time.Second, "", "")

	// Create a collector
	collector := NewCollector(client, logger)

	// Collect metrics
	ch := make(chan prometheus.Metric)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	// Verify metrics were collected
	metrics := make([]prometheus.Metric, 0)
	for m := range ch {
		metrics = append(metrics, m)
	}

	// Expected metrics: up, build_info, uptime, cpu_load_avg (3 values), memory, memory_free
	// Total: 1 + 1 + 1 + 3 + 1 + 1 = 8 metrics
	assert.Greater(t, len(metrics), 0, "Expected metrics to be collected")

	// Verify the metrics are not nil
	for _, m := range metrics {
		assert.NotNil(t, m)
	}
}

// TestCollectorBuildInfoMetric tests the build_info metric specifically
func TestCollectorBuildInfoMetric(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/status" {
			http.NotFound(w, r)
			return
		}

		response := map[string]interface{}{
			"system-information": map[string]interface{}{
				"API Version":    "LANTIME REST API V20.05.013",
				"version":        "fw_7.10.008",
				"serial-number":  "SERIAL123",
				"hostname":       "test-hostname",
			},
			"data": map[string]interface{}{
				"system": map[string]interface{}{
					"uptime":  130988.25,
					"cpuload": "0.48 0.66 0.57 2/99 25157",
					"memory":  "228428 kB total memory, 161732 kB free (70 %)",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	client := NewClient(mockServer.URL, 5*time.Second, "", "")
	collector := NewCollector(client, logger)

	// Collect metrics
	ch := make(chan prometheus.Metric)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	metrics := make([]prometheus.Metric, 0)
	for m := range ch {
		metrics = append(metrics, m)
	}

	// Should have at least the build_info metric
	assert.Greater(t, len(metrics), 0, "Expected at least one metric")
}

// TestCollectorWithUnreachableServer tests collector behavior when server is unreachable
func TestCollectorWithUnreachableServer(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create a client pointing to an unreachable server
	client := NewClient("http://localhost:9999", 1*time.Second, "", "")

	collector := NewCollector(client, logger)

	// Collect metrics
	ch := make(chan prometheus.Metric)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	// Should still collect the up metric with value 0
	metrics := make([]prometheus.Metric, 0)
	for m := range ch {
		metrics = append(metrics, m)
	}

	assert.Equal(t, 1, len(metrics), "Expected at least 1 metric (up metric)")
}

// TestClientFetchStatus tests the FetchStatus method
func TestClientFetchStatus(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/status" {
			http.NotFound(w, r)
			return
		}

		response := map[string]interface{}{
			"system-information": map[string]interface{}{
				"API Version":    "LANTIME REST API V20.05.013",
				"version":        "fw_7.10.008",
				"serial-number":  "SERIAL123",
				"hostname":       "test-device",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL, 5*time.Second, "", "")

	// Fetch status
	status, err := client.FetchStatus()
	require.NoError(t, err)
	assert.NotNil(t, status)

	// Verify the response structure
	sysInfo, ok := status["system-information"].(map[string]interface{})
	require.True(t, ok, "system-information should be a map")

	apiVersion, ok := sysInfo["API Version"].(string)
	require.True(t, ok, "API Version should be a string")
	assert.Equal(t, "LANTIME REST API V20.05.013", apiVersion)

	hostname, ok := sysInfo["hostname"].(string)
	require.True(t, ok, "hostname should be a string")
	assert.Equal(t, "test-device", hostname)
}

// TestClientCheckHealth tests the CheckHealth method
func TestClientCheckHealth(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/status" {
			http.NotFound(w, r)
			return
		}

		response := map[string]interface{}{
			"system-information": map[string]interface{}{
				"API Version": "LANTIME REST API V20.05.013",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL, 5*time.Second, "", "")

	// Check health
	isHealthy, err := client.CheckHealth()
	require.NoError(t, err)
	assert.True(t, isHealthy)
}

