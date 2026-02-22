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
	"log/slog"
	"time"

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
	eventMetric           typedDesc
	storageCapacity       typedDesc
	storageUsed           typedDesc
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
		eventMetric: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_event",
				"Information about events triggered on the Meinberg device",
				[]string{"type", "event"},
				nil,
			),
			valueType: prometheus.CounterValue,
		},
		storageCapacity: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_storage_capacity_bytes",
				"Total size of the storage volume in bytes",
				[]string{"host", "mount"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		storageUsed: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_storage_used_bytes",
				"Used bytes of the storage volume",
				[]string{"host", "mount"},
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
	ch <- c.eventMetric.desc
	ch <- c.storageCapacity.desc
	ch <- c.storageUsed.desc
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

		// Check and parse system-information for build and system info metric
		systemInfoRaw, ok := statusData["system-information"]
		if !ok {
			c.logger.Debug("Key 'system-information' missing in status data")
			return
		}
		systemInfo, ok := systemInfoRaw.(map[string]any)
		if !ok {
			c.logger.Debug("Key 'system-information' is not of expected type map[string]interface{}")
			return
		}

		dataRaw, ok := statusData["data"]
		if !ok {
			c.logger.Debug("Key 'data' missing in status data")
			return
		}
		data, ok := dataRaw.(map[string]any)
		if !ok {
			c.logger.Debug("Key 'data' is not of expected type map[string]interface{}")
			return
		}

		restAPIRaw, ok := data["rest-api"]
		if !ok {
			c.logger.Debug("Key 'rest-api' missing in data")
			return
		}
		restAPI, ok := restAPIRaw.(map[string]any)
		if !ok {
			c.logger.Debug("Key 'rest-api' is not of expected type map[string]interface{}")
			return
		}

		apiVersion, ok := restAPI["api-version"].(string)
		if !ok {
			c.logger.Debug("Key 'api-version' missing or not of type string in rest-api")
			return
		}

		firmwareVersion, ok := systemInfo["version"].(string)
		if !ok {
			c.logger.Debug("Key 'version' missing or not of type string in system-information")
			return
		}
		model, ok := systemInfo["model"].(string)
		if !ok {
			c.logger.Debug("Key 'model' missing or not of type string in system-information")
			return
		}
		serial, ok := systemInfo["serial-number"].(string)
		if !ok {
			c.logger.Debug("Key 'serial-number' missing or not of type string in system-information")
			return
		}
		host, ok = systemInfo["hostname"].(string)
		if !ok {
			c.logger.Debug("Key 'hostname' missing or not of type string in system-information")
			return
		}
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
			load1, load5, load15, err := parseCPULoad(cpuloadStr)
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
			totalBytes, freeBytes, err := parseMemory(memoryStr)
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

		// Parse notification events and emit metrics
		if notifications, ok := data["notification"].(map[string]any); ok {
			if events, ok := notifications["events"].([]any); ok {
				for _, evt := range events {
					event := evt.(map[string]any)
					eventType := event["type"].(string)
					eventName := event["object-id"].(string)
					lastTriggered := event["last-triggered"].(string)

					if lastTriggered != "never" {
						parsedTime, err := time.Parse("2006-01-02T15:04:05", lastTriggered)
						if err != nil {
							c.logger.Debug("Failed to parse 'last-triggered' timestamp", "error", err.Error(), "last-triggered", lastTriggered)
							continue
						}
						ch <- prometheus.MustNewConstMetric(
							c.eventMetric.desc,
							c.eventMetric.valueType,
							float64(parsedTime.Unix()),
							eventType, eventName,
						)
					}
				}
			}
		}

		// Parse and emit storage metrics
		if storageData, ok := system["storage"].([]any); ok {
			for _, rawStorageEntry := range storageData {
				storageEntry, ok := rawStorageEntry.(map[string]any)
				if !ok {
					c.logger.Debug("Failed to parse storage entry", "entry", rawStorageEntry)
					continue
				}

				if mountpoint, ok := storageEntry["mountpoint"].(string); ok {
					if volumeSize, ok := storageEntry["size"].(float64); ok {
						ch <- prometheus.MustNewConstMetric(
							c.storageCapacity.desc,
							c.storageCapacity.valueType,
							volumeSize,
							host, mountpoint,
						)
					}

					if usedBytes, ok := storageEntry["used"].(float64); ok {
						ch <- prometheus.MustNewConstMetric(
							c.storageUsed.desc,
							c.storageUsed.valueType,
							usedBytes,
							host, mountpoint,
						)
					}
				}
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
