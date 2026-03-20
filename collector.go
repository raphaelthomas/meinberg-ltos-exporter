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
	"math"

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
	ntpLeapIndicator      typedDesc
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
				MetricPrefix+"event_last_triggered_seconds",
				"When an event last occurred as seconds since UNIX epoch (0 if never triggered)",
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
				[]string{"host", "refid", "assoc"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		ntpPrecision: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"ntp_precision_seconds",
				"Meinberg NTP precision in seconds",
				[]string{"host", "refid", "assoc"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		ntpRootDelay: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"ntp_root_delay_seconds",
				"Meinberg NTP root delay in seconds",
				[]string{"host", "refid", "assoc"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		ntpRootDispersion: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"ntp_root_dispersion_seconds",
				"Meinberg NTP root dispersion in seconds",
				[]string{"host", "refid", "assoc"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		ntpClockJitter: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"ntp_clock_jitter_seconds",
				"Meinberg NTP clock jitter in seconds",
				[]string{"host", "refid", "assoc"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		ntpClockWander: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"ntp_clock_wander_seconds_per_second",
				"Meinberg NTP clock wander in seconds per second",
				[]string{"host", "refid", "assoc"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		ntpLeapIndicator: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"ntp_leap_indicator",
				"Meinberg NTP leap indicator (0 = no warning, 1 = last minute has 61 seconds, 2 = last minute has 59 seconds, 3 = unknown)",
				[]string{"host", "refid", "assoc"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		ntpLeapSecond: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"ntp_leap_second_seconds",
				"Meinberg NTP leap second (last or next) in seconds since UNIX epoch",
				[]string{"host", "refid", "assoc"},
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
	ch <- c.ntpLeapIndicator.desc
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

	for _, event := range status.Data.Notification.Events {
		ch <- prometheus.MustNewConstMetric(
			c.event.desc,
			c.event.valueType,
			event.LastTriggeredUnix,
			host, event.Type, event.Name,
		)
	}

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

	for _, assoc := range status.Data.NTP {
		if !assoc.IsSys() {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.ntpStratum.desc,
			c.ntpStratum.valueType,
			assoc.Stratum,
			host, assoc.RefID, assoc.Name,
		)
		precisionSeconds := math.Pow(2, assoc.Precision)
		ch <- prometheus.MustNewConstMetric(
			c.ntpPrecision.desc,
			c.ntpPrecision.valueType,
			precisionSeconds,
			host, assoc.RefID, assoc.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ntpRootDelay.desc,
			c.ntpRootDelay.valueType,
			assoc.RootDelay,
			host, assoc.RefID, assoc.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ntpRootDispersion.desc,
			c.ntpRootDispersion.valueType,
			assoc.RootDispersion,
			host, assoc.RefID, assoc.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ntpClockJitter.desc,
			c.ntpClockJitter.valueType,
			assoc.ClockJitter,
			host, assoc.RefID, assoc.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ntpClockWander.desc,
			c.ntpClockWander.valueType,
			assoc.ClockWander,
			host, assoc.RefID, assoc.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ntpLeapIndicator.desc,
			c.ntpLeapIndicator.valueType,
			float64(assoc.LeapIndicator),
			host, assoc.RefID, assoc.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ntpLeapSecond.desc,
			c.ntpLeapSecond.valueType,
			float64(assoc.LeapSecondUnix),
			host, assoc.RefID, assoc.Name,
		)
	}

	for _, slot := range status.Data.Chassis.Slots {
		if slot.Module == nil {
			continue
		}

		if slot.Type == "cpu" {
			ch <- prometheus.MustNewConstMetric(
				c.systemCPUInfo.desc,
				c.systemCPUInfo.valueType,
				1.0,
				host, slot.Module.Info.Model, slot.Module.Info.SerialNumber,
			)
		} else if slot.Type == "clk" {
			oscillatorType := "unknown"
			if slot.Module.SyncStatus != nil {
				oscillatorType = slot.Module.SyncStatus.OscillatorType
			}
			ch <- prometheus.MustNewConstMetric(
				c.receiverInfo.desc,
				c.receiverInfo.valueType,
				1.0,
				host, slot.Name, slot.Module.Info.Model, slot.Module.Info.SerialNumber, slot.Module.Info.SoftwareRevision, oscillatorType,
			)

			if slot.Module.Satellites != nil {
				ch <- prometheus.MustNewConstMetric(
					c.rcvGNSSSatInView.desc,
					c.rcvGNSSSatInView.valueType,
					slot.Module.Satellites.InView,
					host, slot.Name,
				)
				ch <- prometheus.MustNewConstMetric(
					c.rcvGNSSSatGood.desc,
					c.rcvGNSSSatGood.valueType,
					slot.Module.Satellites.Good,
					host, slot.Name,
				)
				ch <- prometheus.MustNewConstMetric(
					c.rcvGNSSLatitude.desc,
					c.rcvGNSSLatitude.valueType,
					slot.Module.Satellites.Latitude,
					host, slot.Name,
				)
				ch <- prometheus.MustNewConstMetric(
					c.rcvGNSSLongitude.desc,
					c.rcvGNSSLongitude.valueType,
					slot.Module.Satellites.Longitude,
					host, slot.Name,
				)
				ch <- prometheus.MustNewConstMetric(
					c.rcvGNSSAltitude.desc,
					c.rcvGNSSAltitude.valueType,
					slot.Module.Satellites.Altitude,
					host, slot.Name,
				)
			}

			if slot.Module.GRC != nil {
				if slot.Module.GRC.Antenna != nil {
					antConnected := 0.0
					if slot.Module.GRC.Antenna.IsConnected {
						antConnected = 1.0
					}
					ch <- prometheus.MustNewConstMetric(
						c.rcvAntConnected.desc,
						c.rcvAntConnected.valueType,
						antConnected,
						host, slot.Name,
					)

					antShortCircuit := 0.0
					if slot.Module.GRC.Antenna.HasShortCircuit {
						antShortCircuit = 1.0
					}
					ch <- prometheus.MustNewConstMetric(
						c.rcvAntShortCircuit.desc,
						c.rcvAntShortCircuit.valueType,
						antShortCircuit,
						host, slot.Name,
					)
				}

				if slot.Module.GRC.Receiver != nil {
					synced := 0.0
					if slot.Module.GRC.Receiver.IsSynchronized {
						synced = 1.0
					}
					ch <- prometheus.MustNewConstMetric(
						c.rcvSynced.desc,
						c.rcvSynced.valueType,
						synced,
						host, slot.Name,
					)

					tracking := 0.0
					if slot.Module.GRC.Receiver.IsTracking {
						tracking = 1.0
					}
					ch <- prometheus.MustNewConstMetric(
						c.rcvTracking.desc,
						c.rcvTracking.valueType,
						tracking,
						host, slot.Name,
					)

					warmBoot := 0.0
					if slot.Module.GRC.Receiver.IsWarmBooting {
						warmBoot = 1.0
					}
					ch <- prometheus.MustNewConstMetric(
						c.rcvWarmBoot.desc,
						c.rcvWarmBoot.valueType,
						warmBoot,
						host, slot.Name,
					)

					coldBoot := 0.0
					if slot.Module.GRC.Receiver.IsColdBooting {
						coldBoot = 1.0
					}
					ch <- prometheus.MustNewConstMetric(
						c.rcvColdBoot.desc,
						c.rcvColdBoot.valueType,
						coldBoot,
						host, slot.Name,
					)
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
