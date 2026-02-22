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
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// typedDesc combines a prometheus.Desc with its value type for cleaner code
type typedDesc struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
}

// Collector implements prometheus.Collector for Meinberg metrics
type Collector struct {
	client *Client
	logger *slog.Logger

	// Metric descriptors
	up                    typedDesc
	buildInfoMetric       typedDesc
	systemInfoMetric      typedDesc
	systemUptimeSeconds   typedDesc
	systemCPULoadAvg      typedDesc
	systemMemoryBytes     typedDesc
	systemMemoryFreeBytes typedDesc
}

// NewCollector creates a new Meinberg collector
func NewCollector(client *Client, logger *slog.Logger) *Collector {
	return &Collector{
		client: client,
		logger: logger,
		up: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_up",
				"Indicates if the Meinberg LTOS device is reachable (1 = up, 0 = down)",
				[]string{"host", "target"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		buildInfoMetric: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_build_info",
				"Meinberg device build information as labels (e.g., API version, firmware version, host)",
				[]string{"api_version", "firmware_version", "host"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		systemInfoMetric: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_system_info",
				"Meinberg system information as labels (e.g., model, serial number, host)",
				[]string{"model", "serial_number", "host"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		systemUptimeSeconds: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_system_uptime_seconds",
				"System uptime in seconds",
				[]string{"host"},
				nil,
			),
			valueType: prometheus.CounterValue,
		},
		systemCPULoadAvg: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_system_cpu_load_avg",
				"CPU load averaged over 1, 5, and 15 minutes",
				[]string{"host", "period"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		systemMemoryBytes: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_system_memory_bytes",
				"Total memory in bytes",
				[]string{"host"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		systemMemoryFreeBytes: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_system_memory_free_bytes",
				"Free memory in bytes",
				[]string{"host"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
	}
}

// Describe implements prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up.desc
	ch <- c.buildInfoMetric.desc
	ch <- c.systemInfoMetric.desc
	ch <- c.systemUptimeSeconds.desc
	ch <- c.systemCPULoadAvg.desc
	ch <- c.systemMemoryBytes.desc
	ch <- c.systemMemoryFreeBytes.desc
}

// parseCPULoad parses the cpuload string and returns the 1, 5, and 15 minute averages
// Example input: "0.48 0.66 0.57 2/99 25157"
func (c *Collector) parseCPULoad(cpuloadStr string) (float64, float64, float64, error) {
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
func (c *Collector) parseMemory(memoryStr string) (float64, float64, error) {
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

// Collect implements prometheus.Collector
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	host := "unknown"
	upValue := 0.0
	statusData, err := c.client.FetchStatus()
	if err != nil {
		c.logger.Debug("Failed to fetch status data", "error", err.Error())
	} else {
		upValue = 1.0

		// Parse system-information for build and system info metric
		systemInfo := statusData["system-information"].(map[string]any)
		apiVersion := systemInfo["API Version"].(string)
		firmwareVersion := systemInfo["version"].(string)
		model := systemInfo["model"].(string)
		serial := systemInfo["serial-number"].(string)
		host = systemInfo["hostname"].(string)

		// Send the build info metric
		ch <- prometheus.MustNewConstMetric(
			c.buildInfoMetric.desc,
			c.buildInfoMetric.valueType,
			1.0,
			apiVersion, firmwareVersion, host,
		)

		// Send the system info metric
		ch <- prometheus.MustNewConstMetric(
			c.systemInfoMetric.desc,
			c.systemInfoMetric.valueType,
			1.0,
			model, serial, host,
		)

		// Parse system data for system information metrics
		data := statusData["data"].(map[string]any)
		system := data["system"].(map[string]any)

		// Extract uptime (already in seconds)
		if uptime, ok := system["uptime"].(float64); ok {
			ch <- prometheus.MustNewConstMetric(
				c.systemUptimeSeconds.desc,
				c.systemUptimeSeconds.valueType,
				uptime,
				host,
			)
		}

		// Extract and parse CPU load averages
		if cpuloadStr, ok := system["cpuload"].(string); ok {
			load1, load5, load15, err := c.parseCPULoad(cpuloadStr)
			if err != nil {
				c.logger.Debug("Failed to parse CPU load", "error", err.Error())
			} else {
				// Send 1-minute average
				ch <- prometheus.MustNewConstMetric(
					c.systemCPULoadAvg.desc,
					c.systemCPULoadAvg.valueType,
					load1,
					host, "1",
				)
				// Send 5-minute average
				ch <- prometheus.MustNewConstMetric(
					c.systemCPULoadAvg.desc,
					c.systemCPULoadAvg.valueType,
					load5,
					host, "5",
				)
				// Send 15-minute average
				ch <- prometheus.MustNewConstMetric(
					c.systemCPULoadAvg.desc,
					c.systemCPULoadAvg.valueType,
					load15,
					host, "15",
				)
			}
		}

		// Extract and parse memory information
		if memoryStr, ok := system["memory"].(string); ok {
			totalBytes, freeBytes, err := c.parseMemory(memoryStr)
			if err == nil {
				ch <- prometheus.MustNewConstMetric(
					c.systemMemoryBytes.desc,
					c.systemMemoryBytes.valueType,
					totalBytes,
					host,
				)
				ch <- prometheus.MustNewConstMetric(
					c.systemMemoryFreeBytes.desc,
					c.systemMemoryFreeBytes.valueType,
					freeBytes,
					host,
				)
			} else {
				c.logger.Debug("Failed to parse memory", "error", err.Error())
			}
		}
	}

	// Create and send the up metric
	ch <- prometheus.MustNewConstMetric(
		c.up.desc,
		c.up.valueType,
		upValue,
		host, c.client.Target(),
	)
}

// Register registers the collector with Prometheus
func (c *Collector) Register() error {
	return prometheus.Register(c)
}
