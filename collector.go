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
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const MetricPrefix = "meinberg_ltos_"

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
	buildInfo             typedDesc
	systemInfo            typedDesc
	systemUptimeSeconds   typedDesc
	systemCPUInfo         typedDesc
	systemCPULoadAvg      typedDesc
	systemMemoryBytes     typedDesc
	systemMemoryFreeBytes typedDesc
	event                 typedDesc
	storageTotal          typedDesc
	storageUsed           typedDesc
	receiverInfo          typedDesc
	rcvGNSSSatInView      typedDesc
	rcvGNSSSatGood        typedDesc
	rcvGNSSLatitude       typedDesc
	rcvGNSSLongitude      typedDesc
	rcvGNSSAltitude       typedDesc
	rcvAntConnected       typedDesc
	rcvAntShortCircuit    typedDesc
	rcvSynced             typedDesc
	rcvTracking           typedDesc
	rcvColdBoot           typedDesc
	rcvWarmBoot           typedDesc
	ntpStratum            typedDesc
	ntpPrecision          typedDesc
	ntpRootDelay          typedDesc
	ntpRootDispersion     typedDesc
	ntpClockJitter        typedDesc
	ntpClockWander        typedDesc
	ntpLeapAnnounced      typedDesc
	ntpLeapSecond         typedDesc
}

// NewCollector creates a new Meinberg collector
func NewCollector(client *Client, logger *slog.Logger) *Collector {
	return &Collector{
		client: client,
		logger: logger,
		up: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"up",
				"Indicates if the Meinberg LTOS device is reachable (1 = up, 0 = down)",
				[]string{"host", "target"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		buildInfo: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"build_info",
				"Meinberg device build information as labels (e.g., API version, firmware version, host)",
				[]string{"host", "api_version", "firmware_version"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		systemInfo: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"system_info",
				"Meinberg system information as labels (e.g., model, serial number, host)",
				[]string{"host", "model", "serial_number"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		systemUptimeSeconds: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"system_uptime_seconds",
				"System uptime in seconds",
				[]string{"host"},
				nil,
			),
			valueType: prometheus.CounterValue,
		},
		systemCPUInfo: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"system_cpu_info",
				"CPU information as labels (model, serial, etc.)",
				[]string{"host", "model", "serial_number"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		systemCPULoadAvg: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"system_cpu_load_avg",
				"CPU load averaged over 1, 5, and 15 minutes",
				[]string{"host", "period"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		systemMemoryBytes: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"system_memory_bytes",
				"Total memory in bytes",
				[]string{"host"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		systemMemoryFreeBytes: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"system_memory_free_bytes",
				"Free memory in bytes",
				[]string{"host"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		event: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"event",
				"Information about events triggered on the Meinberg device",
				[]string{"host", "type", "event"},
				nil,
			),
			valueType: prometheus.CounterValue,
		},
		storageTotal: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"storage_total_bytes",
				"Total size of the storage volume in bytes",
				[]string{"host", "mount"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		storageUsed: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"storage_used_bytes",
				"Used bytes of the storage volume",
				[]string{"host", "mount"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		receiverInfo: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"receiver_info",
				"Meinberg receiver module information as labels (model, serial number, software revision, oscillator type)",
				[]string{"host", "slot_id", "model", "serial_number", "software_revision", "oscillator_type"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		rcvGNSSSatInView: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"receiver_gnss_satellites_in_view",
				"Meinberg GNSS receiver satellites in view",
				[]string{"host", "slot_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		rcvGNSSSatGood: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"receiver_gnss_satellites_good",
				"Meinberg GNSS receiver good satellites",
				[]string{"host", "slot_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		rcvGNSSLatitude: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"receiver_gnss_latitude_degrees",
				"Meinberg GNSS receiver latitude",
				[]string{"host", "slot_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		rcvGNSSLongitude: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"receiver_gnss_longitude_degrees",
				"Meinberg GNSS receiver longitude",
				[]string{"host", "slot_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		rcvGNSSAltitude: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"receiver_gnss_altitude_meters",
				"Meinberg GNSS receiver altitude",
				[]string{"host", "slot_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		rcvAntConnected: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"receiver_gnss_antenna_connected",
				"Meinberg GNSS receiver antenna connected (1 = connected, 0 = not connected)",
				[]string{"host", "slot_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		rcvAntShortCircuit: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"receiver_gnss_antenna_short_circuit",
				"Meinberg GNSS receiver antenna short circuit detected (1 = short circuit, 0 = no short circuit)",
				[]string{"host", "slot_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		rcvSynced: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"receiver_synced",
				"Meinberg receiver synchronization status (1 = synced, 0 = not synced)",
				[]string{"host", "slot_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		rcvTracking: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"receiver_tracking",
				"Meinberg receiver tracking status (1 = tracking, 0 = not tracking)",
				[]string{"host", "slot_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		rcvColdBoot: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"receiver_cold_boot",
				"Meinberg receiver cold boot status (1 = cold boot, 0 = not cold boot)",
				[]string{"host", "slot_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		rcvWarmBoot: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"receiver_warm_boot",
				"Meinberg receiver warm boot status (1 = warm boot, 0 = not warm boot)",
				[]string{"host", "slot_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		ntpStratum: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"ntp_stratum",
				"Meinberg NTP stratum level",
				[]string{"host"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		ntpPrecision: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"ntp_precision_seconds",
				"Meinberg NTP precision in seconds",
				[]string{"host"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		ntpRootDelay: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"ntp_root_delay_seconds",
				"Meinberg NTP root delay in seconds",
				[]string{"host"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		ntpRootDispersion: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"ntp_root_dispersion_seconds",
				"Meinberg NTP root dispersion in seconds",
				[]string{"host"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		ntpClockJitter: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"ntp_clock_jitter_seconds",
				"Meinberg NTP clock jitter in seconds",
				[]string{"host"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		ntpClockWander: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"ntp_clock_wander_seconds_per_second",
				"Meinberg NTP clock wander in seconds per second",
				[]string{"host"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		ntpLeapAnnounced: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"ntp_leap_announced",
				"Meinberg NTP leap second announced status (1 = leap second announced, 0 = no leap second announced)",
				[]string{"host"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		ntpLeapSecond: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"ntp_leap_second",
				"Meinberg NTP leap second (last or next) in seconds since epoch",
				[]string{"host"},
				nil,
			),
			valueType: prometheus.CounterValue,
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
	ch <- c.storageTotal.desc
	ch <- c.storageUsed.desc
	ch <- c.receiverInfo.desc
	ch <- c.rcvGNSSSatInView.desc
	ch <- c.rcvGNSSSatGood.desc
	ch <- c.rcvGNSSLatitude.desc
	ch <- c.rcvGNSSLongitude.desc
	ch <- c.rcvGNSSAltitude.desc
	ch <- c.rcvAntConnected.desc
	ch <- c.rcvAntShortCircuit.desc
	ch <- c.rcvSynced.desc
	ch <- c.rcvTracking.desc
	ch <- c.rcvColdBoot.desc
	ch <- c.rcvWarmBoot.desc
	ch <- c.ntpStratum.desc
	ch <- c.ntpPrecision.desc
	ch <- c.ntpRootDelay.desc
	ch <- c.ntpRootDispersion.desc
	ch <- c.ntpClockJitter.desc
	ch <- c.ntpClockWander.desc
	ch <- c.ntpLeapAnnounced.desc
	ch <- c.ntpLeapSecond.desc
}

// Collect implements prometheus.Collector
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Debug("Collecting metrics from Meinberg LTOS device", "target", c.client.Target())

	host := "unknown"
	up := 0.0

	defer func() {
		ch <- prometheus.MustNewConstMetric(
			c.up.desc,
			c.up.valueType,
			up,
			host, c.client.Target(),
		)
	}()

	status, err := c.client.FetchStatus()
	if err != nil {
		c.logger.Warn("Failed to fetch Meinberg LTOS device status", "error", err.Error())
		return
	}

	up = 1.0
	host = status.SystemInformation.Hostname

	ch <- prometheus.MustNewConstMetric(
		c.buildInfo.desc,
		c.buildInfo.valueType,
		1.0,
		host, status.Data.RestAPI.Version, status.SystemInformation.Version,
	)
	ch <- prometheus.MustNewConstMetric(
		c.systemInfo.desc,
		c.systemInfo.valueType,
		1.0,
		host, status.SystemInformation.Model, status.SystemInformation.SerialNumber,
	)
	ch <- prometheus.MustNewConstMetric(
		c.systemUptimeSeconds.desc,
		c.systemUptimeSeconds.valueType,
		status.Data.System.UptimeSeconds,
		host,
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemCPULoadAvg.desc,
		c.systemCPULoadAvg.valueType,
		status.Data.System.CPULoad.Load1,
		host, "1",
	)
	ch <- prometheus.MustNewConstMetric(
		c.systemCPULoadAvg.desc,
		c.systemCPULoadAvg.valueType,
		status.Data.System.CPULoad.Load5,
		host, "5",
	)
	ch <- prometheus.MustNewConstMetric(
		c.systemCPULoadAvg.desc,
		c.systemCPULoadAvg.valueType,
		status.Data.System.CPULoad.Load15,
		host, "15",
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemMemoryBytes.desc,
		c.systemMemoryBytes.valueType,
		status.Data.System.Memory.Total,
		host,
	)
	ch <- prometheus.MustNewConstMetric(
		c.systemMemoryFreeBytes.desc,
		c.systemMemoryFreeBytes.valueType,
		status.Data.System.Memory.Free,
		host,
	)

	// Parse notification events and emit metrics
	notifications := status.Data.Notification
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

	// Parse and emit storage metrics
	for _, mount := range status.Data.System.Mounts {
		ch <- prometheus.MustNewConstMetric(
			c.storageTotal.desc,
			c.storageTotal.valueType,
			mount.Size,
			host, mount.Mountpoint,
		)
		ch <- prometheus.MustNewConstMetric(
			c.storageUsed.desc,
			c.storageUsed.valueType,
			mount.Used,
			host, mount.Mountpoint,
		)
	}

	// Parse and emit NTP service metrics
	ntpData := status.Data.NTP
	for _, assoc := range ntpData {
		if assoc["object-id"].(string) != "sys" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.ntpStratum.desc,
			c.ntpStratum.valueType,
			assoc["stratum"].(float64),
			host,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ntpPrecision.desc,
			c.ntpPrecision.valueType,
			assoc["precision"].(float64),
			host,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ntpRootDelay.desc,
			c.ntpRootDelay.valueType,
			assoc["rootdelay"].(float64),
			host,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ntpRootDispersion.desc,
			c.ntpRootDispersion.valueType,
			assoc["rootdisp"].(float64),
			host,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ntpClockJitter.desc,
			c.ntpClockJitter.valueType,
			assoc["clk-jitter"].(float64),
			host,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ntpClockWander.desc,
			c.ntpClockWander.valueType,
			assoc["clk-wander"].(float64),
			host,
		)
		var leap float64
		switch leapRaw := assoc["leap"].(type) {
		case float64:
			leap = leapRaw
		case string:
			leap, _ = strconv.ParseFloat(leapRaw, 64)
		}
		ch <- prometheus.MustNewConstMetric(
			c.ntpLeapAnnounced.desc,
			c.ntpLeapAnnounced.valueType,
			leap,
			host,
		)
		// TODO convert weird timestamp into epoch
		leapSecRaw := assoc["leapsec"].(string)
		leapSecTime, _ := time.Parse("200601021504", leapSecRaw)
		ch <- prometheus.MustNewConstMetric(
			c.ntpLeapSecond.desc,
			c.ntpLeapSecond.valueType,
			float64(leapSecTime.Unix()),
			host,
		)
	}

	// Parse and emit metrics from chassis slots
	chassisData := status.Data.Chassis0
	if slots, ok := chassisData["slots"].([]any); ok {
		for _, slotRaw := range slots {
			slot := slotRaw.(map[string]any)
			slotType := slot["slot-type"].(string)
			slotID := slot["slot-id"].(string)

			if slotType == "cpu" {
				if cpuModuleData, ok := slot["module"].(map[string]any); ok {
					if cpuInfoData := cpuModuleData["info"].(map[string]any); ok {
						cpuModel := cpuInfoData["model"].(string)
						cpuSerial := cpuInfoData["serial-number"].(string)
						ch <- prometheus.MustNewConstMetric(
							c.systemCPUInfo.desc,
							c.systemCPUInfo.valueType,
							1.0,
							host, cpuModel, cpuSerial,
						)
					}
				}
			} else if slotType == "clk" {
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
							c.receiverInfo.desc,
							c.receiverInfo.valueType,
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
							c.rcvGNSSSatInView.desc,
							c.rcvGNSSSatInView.valueType,
							satInView,
							host, slotID,
						)
						ch <- prometheus.MustNewConstMetric(
							c.rcvGNSSSatGood.desc,
							c.rcvGNSSSatGood.valueType,
							satGood,
							host, slotID,
						)
						ch <- prometheus.MustNewConstMetric(
							c.rcvGNSSLatitude.desc,
							c.rcvGNSSLatitude.valueType,
							lat,
							host, slotID,
						)
						ch <- prometheus.MustNewConstMetric(
							c.rcvGNSSLongitude.desc,
							c.rcvGNSSLongitude.valueType,
							lon,
							host, slotID,
						)
						ch <- prometheus.MustNewConstMetric(
							c.rcvGNSSAltitude.desc,
							c.rcvGNSSAltitude.valueType,
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
								c.rcvAntConnected.desc,
								c.rcvAntConnected.valueType,
								antConnected,
								host, slotID,
							)

							antShortCircuit := 0.0
							if grcAntData["short-circuit"].(bool) {
								antShortCircuit = 1.0
							}
							ch <- prometheus.MustNewConstMetric(
								c.rcvAntShortCircuit.desc,
								c.rcvAntShortCircuit.valueType,
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
								c.rcvSynced.desc,
								c.rcvSynced.valueType,
								synced,
								host, slotID,
							)

							tracking := 0.0
							if grcRcvData["tracking"].(bool) {
								tracking = 1.0
							}
							ch <- prometheus.MustNewConstMetric(
								c.rcvTracking.desc,
								c.rcvTracking.valueType,
								tracking,
								host, slotID,
							)

							warmBoot := 0.0
							if grcRcvData["warm-boot"].(bool) {
								warmBoot = 1.0
							}
							ch <- prometheus.MustNewConstMetric(
								c.rcvWarmBoot.desc,
								c.rcvWarmBoot.valueType,
								warmBoot,
								host, slotID,
							)

							coldBoot := 0.0
							if grcRcvData["cold-boot"].(bool) {
								coldBoot = 1.0
							}
							ch <- prometheus.MustNewConstMetric(
								c.rcvColdBoot.desc,
								c.rcvColdBoot.valueType,
								coldBoot,
								host, slotID,
							)
						}
					}
				}
			}
		}
	}

	c.logger.Debug("Done collecting metrics from Meinberg LTOS device", "target", c.client.Target())
}

// Register registers the collector with Prometheus
func (c *Collector) Register() error {
	return prometheus.Register(c)
}
