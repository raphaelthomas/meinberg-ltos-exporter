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
	up                      typedDesc
	buildInfo               typedDesc
	systemInfo              typedDesc
	systemUptimeSeconds     typedDesc
	systemCPULoadAvg        typedDesc
	systemMemoryBytes       typedDesc
	systemMemoryFreeBytes   typedDesc
	event                   typedDesc
	storageCapacity         typedDesc
	storageUsed             typedDesc
	clockInfo               typedDesc
	clockRcvGNSSSatInView   typedDesc
	clockRcvGNSSSatGood     typedDesc
	clockRcvGNSSLatitude    typedDesc
	clockRcvGNSSLongitude   typedDesc
	clockRcvGNSSAltitude    typedDesc
	clockRcvAntConnected    typedDesc
	clockRcvAntShortCircuit typedDesc
	clockRcvSynced          typedDesc
	clockRcvTracking        typedDesc
	clockRcvColdBoot        typedDesc
	clockRcvWarmBoot        typedDesc
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
		buildInfo: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_build_info",
				"Meinberg device build information as labels (e.g., API version, firmware version, host)",
				[]string{"host", "api_version", "firmware_version"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		systemInfo: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_system_info",
				"Meinberg system information as labels (e.g., model, serial number, host)",
				[]string{"host", "model", "serial_number"},
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
		event: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_event",
				"Information about events triggered on the Meinberg device",
				[]string{"host", "type", "event"},
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
		clockInfo: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_clock_info",
				"Meinberg clock module information as labels (e.g., clock ID, model, serial number, softwware revision)",
				[]string{"host", "clock_id", "model", "serial_number", "software_revision", "oscillator_type"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clockRcvGNSSSatInView: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_clock_rcv_gnss_satellites_in_view",
				"Meinberg clock GNSS receiver satellites in view",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clockRcvGNSSSatGood: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_clock_rcv_gnss_satellites_good",
				"Meinberg clock GNSS receiver good satellites",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clockRcvGNSSLatitude: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_clock_rcv_gnss_latitude_degrees",
				"Meinberg clock GNSS receiver latitude",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clockRcvGNSSLongitude: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_clock_rcv_gnss_longitude_degrees",
				"Meinberg clock GNSS receiver longitude",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clockRcvGNSSAltitude: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_clock_rcv_gnss_altitude_meters",
				"Meinberg clock GNSS receiver altitude",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clockRcvAntConnected: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_clock_gnss_antenna_connected",
				"Meinberg clock GNSS receiver antenna connected (1 = connected, 0 = not connected)",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clockRcvAntShortCircuit: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_clock_gnss_antenna_short_circuit",
				"Meinberg clock GNSS receiver antenna short circuit detected (1 = short circuit, 0 = no short circuit)",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clockRcvSynced: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_clock_rcv_synced",
				"Meinberg clock receiver synchronization status (1 = synced, 0 = not synced)",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clockRcvTracking: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_clock_rcv_tracking",
				"Meinberg clock receiver tracking status (1 = tracking, 0 = not tracking)",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clockRcvColdBoot: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_clock_rcv_cold_boot",
				"Meinberg clock receiver cold boot status (1 = cold boot, 0 = not cold boot)",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clockRcvWarmBoot: typedDesc{
			desc: prometheus.NewDesc(
				"mbg_ltos_clock_rcv_warm_boot",
				"Meinberg clock receiver warm boot status (1 = warm boot, 0 = not warm boot)",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
	}
}

// Describe implements prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up.desc
	ch <- c.buildInfo.desc
	ch <- c.systemInfo.desc
	ch <- c.systemUptimeSeconds.desc
	ch <- c.systemCPULoadAvg.desc
	ch <- c.systemMemoryBytes.desc
	ch <- c.systemMemoryFreeBytes.desc
	ch <- c.event.desc
	ch <- c.storageCapacity.desc
	ch <- c.storageUsed.desc
	ch <- c.clockInfo.desc
	ch <- c.clockRcvGNSSSatInView.desc
	ch <- c.clockRcvGNSSSatGood.desc
	ch <- c.clockRcvGNSSLatitude.desc
	ch <- c.clockRcvGNSSLongitude.desc
	ch <- c.clockRcvGNSSAltitude.desc
	ch <- c.clockRcvAntConnected.desc
	ch <- c.clockRcvAntShortCircuit.desc
	ch <- c.clockRcvSynced.desc
	ch <- c.clockRcvTracking.desc
	ch <- c.clockRcvColdBoot.desc
	ch <- c.clockRcvWarmBoot.desc
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
			c.buildInfo.desc,
			c.buildInfo.valueType,
			1.0,
			host, apiVersion, firmwareVersion,
		)

		// Send the system info metric
		ch <- prometheus.MustNewConstMetric(
			c.systemInfo.desc,
			c.systemInfo.valueType,
			1.0,
			host, model, serial,
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
							c.event.desc,
							c.event.valueType,
							float64(parsedTime.Unix()),
							host, eventType, eventName,
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

		// Parse and emit clock information metrics
		// TODO generalize this to "slot information metrics (pwr, cpu)"
		if chassisData, ok := data["chassis0"].(map[string]any); ok {
			if slots, ok := chassisData["slots"].([]any); ok {
				for _, slotRaw := range slots {
					slot := slotRaw.(map[string]any)
					slotType := slot["slot-type"].(string)
					slotID := slot["slot-id"].(string)

					if slotType == "clk" {
						if moduleData, ok := slot["module"].(map[string]any); ok {
							if moduleInfoData, ok := moduleData["info"].(map[string]any); ok {
								model := moduleInfoData["model"].(string)
								serial := moduleInfoData["serial-number"].(string)
								softwareRevision := moduleInfoData["software-revision"].(string)
								oscillatorType := "unknown"
								if syncStatus, ok := moduleData["sync-status"].(map[string]any); ok {
									oscillatorType = syncStatus["osc-type"].(string)
								}
								ch <- prometheus.MustNewConstMetric(
									c.clockInfo.desc,
									c.clockInfo.valueType,
									1.0,
									host, slotID, model, serial, softwareRevision, oscillatorType,
								)
							}

							if satellitesData, ok := moduleData["satellites"].(map[string]any); ok {
								satInView := satellitesData["satellites-in-view"].(float64)
								satGood := satellitesData["good-satellites"].(float64)
								lat := satellitesData["latitude"].(float64)
								lon := satellitesData["longitude"].(float64)
								alt := satellitesData["altitude"].(float64)

								ch <- prometheus.MustNewConstMetric(
									c.clockRcvGNSSSatInView.desc,
									c.clockRcvGNSSSatInView.valueType,
									satInView,
									host, slotID,
								)
								ch <- prometheus.MustNewConstMetric(
									c.clockRcvGNSSSatGood.desc,
									c.clockRcvGNSSSatGood.valueType,
									satGood,
									host, slotID,
								)
								ch <- prometheus.MustNewConstMetric(
									c.clockRcvGNSSLatitude.desc,
									c.clockRcvGNSSLatitude.valueType,
									lat,
									host, slotID,
								)
								ch <- prometheus.MustNewConstMetric(
									c.clockRcvGNSSLongitude.desc,
									c.clockRcvGNSSLongitude.valueType,
									lon,
									host, slotID,
								)
								ch <- prometheus.MustNewConstMetric(
									c.clockRcvGNSSAltitude.desc,
									c.clockRcvGNSSAltitude.valueType,
									alt,
									host, slotID,
								)
							}

							if grcData, ok := moduleData["grc"].(map[string]any); ok {
								if grcAntData, ok := grcData["antenna"].(map[string]any); ok {
									antConnected := 0.0
									if grcAntData["connected"].(bool) {
										antConnected = 1.0
									}
									ch <- prometheus.MustNewConstMetric(
										c.clockRcvAntConnected.desc,
										c.clockRcvAntConnected.valueType,
										antConnected,
										host, slotID,
									)

									antShortCircuit := 0.0
									if grcAntData["short-circuit"].(bool) {
										antShortCircuit = 1.0
									}
									ch <- prometheus.MustNewConstMetric(
										c.clockRcvAntShortCircuit.desc,
										c.clockRcvAntShortCircuit.valueType,
										antShortCircuit,
										host, slotID,
									)
								}
								if grcRcvData, ok := grcData["receiver"].(map[string]any); ok {
									synced := 0.0
									if grcRcvData["synchronized"].(bool) {
										synced = 1.0
									}
									ch <- prometheus.MustNewConstMetric(
										c.clockRcvSynced.desc,
										c.clockRcvSynced.valueType,
										synced,
										host, slotID,
									)

									tracking := 0.0
									if grcRcvData["tracking"].(bool) {
										tracking = 1.0
									}
									ch <- prometheus.MustNewConstMetric(
										c.clockRcvTracking.desc,
										c.clockRcvTracking.valueType,
										tracking,
										host, slotID,
									)

									warmBoot := 0.0
									if grcRcvData["warm-boot"].(bool) {
										warmBoot = 1.0
									}
									ch <- prometheus.MustNewConstMetric(
										c.clockRcvWarmBoot.desc,
										c.clockRcvWarmBoot.valueType,
										warmBoot,
										host, slotID,
									)

									coldBoot := 0.0
									if grcRcvData["cold-boot"].(bool) {
										coldBoot = 1.0
									}
									ch <- prometheus.MustNewConstMetric(
										c.clockRcvColdBoot.desc,
										c.clockRcvColdBoot.valueType,
										coldBoot,
										host, slotID,
									)
								}
							}
						}
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
