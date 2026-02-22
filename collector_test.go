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
				"version":       "fw_7.10.008",
				"serial-number": "0123456789",
				"hostname":      "mbg1.time.example.com",
				"model":         "M600",
			},
			"data": map[string]interface{}{
				"object-id": "status",
				"rest-api": map[string]interface{}{
					"api-version": "20.05.013",
				},
				"system": map[string]interface{}{
					"uptime": 130988.25,
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

	metrics := make([]prometheus.Metric, 0)
	for m := range ch {
		metrics = append(metrics, m)
	}

	// Expected metrics: up, build_info, uptime, cpu_load_avg (3 values), memory, memory_free
	assert.Greater(t, len(metrics), 0, "Expected metrics to be collected")

	for _, m := range metrics {
		assert.NotNil(t, m)
	}
}
