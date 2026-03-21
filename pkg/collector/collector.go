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

// Package collector implements the Prometheus collector for Meinberg LTOS metrics
package collector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/ltosapi"
)

const MetricPrefix = "meinberg_ltos_"

// typedDesc combines a prometheus.Desc with its value type for cleaner code
type typedDesc struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
}

func (td typedDesc) mustNewConstMetric(value float64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(td.desc, td.valueType, value, labels...)
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

// Collector implements prometheus.Collector for Meinberg metrics
type Collector struct {
	client *ltosapi.Client
	logger *slog.Logger

	up                        typedDesc
	clkRcvGNSSSatInView       typedDesc
	clkRcvGNSSSatGood         typedDesc
	clkRcvGNSSLatitude        typedDesc
	clkRcvGNSSLongitude       typedDesc
	clkRcvGNSSAltitude        typedDesc
	clkRcvGNSSAntConnected    typedDesc
	clkRcvGNSSAntShortCircuit typedDesc
	clkRcvGNSSSynced          typedDesc
	clkRcvGNSSTracking        typedDesc
	clkRcvGNSSColdBoot        typedDesc
	clkRcvGNSSWarmBoot        typedDesc
	clkRcvDCF77FieldStrength  typedDesc
	clkRcvDCF77Correlation    typedDesc
}

// NewCollector creates a new Meinberg collector
func NewCollector(client *ltosapi.Client, logger *slog.Logger) *Collector {
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
		clkRcvGNSSSatInView: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"clock_receiver_gnss_satellites_in_view",
				"Number of satellites (theoretically) in view of the GNSS receiver",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clkRcvGNSSSatGood: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"clock_receiver_gnss_satellites_good",
				"Number of good satellites for the GNSS receiver",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clkRcvGNSSLatitude: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"clock_receiver_gnss_latitude_degrees",
				"Meinberg GNSS receiver latitude",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clkRcvGNSSLongitude: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"clock_receiver_gnss_longitude_degrees",
				"Meinberg GNSS receiver longitude",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clkRcvGNSSAltitude: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"clock_receiver_gnss_altitude_meters",
				"Meinberg GNSS receiver altitude",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clkRcvGNSSAntConnected: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"clock_receiver_gnss_antenna_connected",
				"Meinberg GNSS receiver antenna connected (1 = connected, 0 = not connected)",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clkRcvGNSSAntShortCircuit: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"clock_receiver_gnss_antenna_short_circuit",
				"Meinberg GNSS receiver antenna short circuit detected (1 = short circuit, 0 = no short circuit)",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clkRcvGNSSSynced: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"clock_receiver_gnss_synchronized",
				"Meinberg GNSS receiver synchronization status (1 = synced, 0 = not synced)",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clkRcvGNSSTracking: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"clock_receiver_gnss_tracking",
				"Meinberg GNSS receiver tracking status (1 = tracking, 0 = not tracking)",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clkRcvGNSSColdBoot: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"clock_receiver_gnss_cold_boot",
				"GNSS receiver cold boot status (1 = cold boot, 0 = not cold boot)",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clkRcvGNSSWarmBoot: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"clock_receiver_gnss_warm_boot",
				"GNSS receiver warm boot status (1 = warm boot, 0 = not warm boot)",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clkRcvDCF77FieldStrength: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"clock_receiver_dcf77_field_strength",
				"DCF77 receiver field strength",
				[]string{"host", "clock_id"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		clkRcvDCF77Correlation: typedDesc{
			desc: prometheus.NewDesc(
				MetricPrefix+"clock_receiver_dcf77_correlation",
				"DCF77 receiver correlation",
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
	describeSystem(ch)
	describeEvent(ch)
	describeStorage(ch)
	describeClock(ch)
	ch <- c.clkRcvGNSSSatInView.desc
	ch <- c.clkRcvGNSSSatGood.desc
	ch <- c.clkRcvGNSSLatitude.desc
	ch <- c.clkRcvGNSSLongitude.desc
	ch <- c.clkRcvGNSSAltitude.desc
	ch <- c.clkRcvGNSSAntConnected.desc
	ch <- c.clkRcvGNSSAntShortCircuit.desc
	ch <- c.clkRcvGNSSSynced.desc
	ch <- c.clkRcvGNSSTracking.desc
	ch <- c.clkRcvGNSSColdBoot.desc
	ch <- c.clkRcvGNSSWarmBoot.desc
	ch <- c.clkRcvDCF77FieldStrength.desc
	ch <- c.clkRcvDCF77Correlation.desc
	describeNTP(ch)
}

// Collect implements prometheus.Collector
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Debug("Collecting metrics from Meinberg LTOS device", "target", c.client.Target())

	host := "unknown"
	up := 0.0

	defer func() {
		ch <- c.up.mustNewConstMetric(up, host, c.client.Target())
	}()

	status, err := c.client.FetchStatus()
	if err != nil {
		c.logger.Warn("Failed to fetch Meinberg LTOS device status", "error", err.Error())
		return
	}

	up = 1.0
	host = status.SystemInformation.Hostname

	c.collectSystem(ch, host, status.SystemInformation, status.Data.System, status.Data.RestAPI, status.Data.Chassis.Slots)
	c.collectEvent(ch, host, status.Data.Notification.Events)
	c.collectStorage(ch, host, status.Data.System.Mounts)
	c.collectNTP(ch, host, status.Data.NTP)
	c.collectClock(ch, host, status.Data.Chassis.Slots)

	for _, slot := range status.Data.Chassis.Slots {
		if slot.Module == nil || slot.Type != "clk" {
			continue
		}

		if slot.Module.Satellites != nil {
			ch <- c.clkRcvGNSSSatInView.mustNewConstMetric(slot.Module.Satellites.InView, host, slot.Name)
			ch <- c.clkRcvGNSSSatGood.mustNewConstMetric(slot.Module.Satellites.Good, host, slot.Name)
			ch <- c.clkRcvGNSSLatitude.mustNewConstMetric(slot.Module.Satellites.Latitude, host, slot.Name)
			ch <- c.clkRcvGNSSLongitude.mustNewConstMetric(slot.Module.Satellites.Longitude, host, slot.Name)
			ch <- c.clkRcvGNSSAltitude.mustNewConstMetric(slot.Module.Satellites.Altitude, host, slot.Name)
		}

		if slot.Module.GRC != nil {
			if slot.Module.GRC.Antenna != nil {
				ch <- c.clkRcvGNSSAntConnected.mustNewConstMetric(boolToFloat64(slot.Module.GRC.Antenna.IsConnected), host, slot.Name)
				ch <- c.clkRcvGNSSAntShortCircuit.mustNewConstMetric(boolToFloat64(slot.Module.GRC.Antenna.HasShortCircuit), host, slot.Name)
			}

			if slot.Module.GRC.Receiver != nil {
				ch <- c.clkRcvGNSSSynced.mustNewConstMetric(boolToFloat64(slot.Module.GRC.Receiver.IsSynchronized), host, slot.Name)
				ch <- c.clkRcvGNSSTracking.mustNewConstMetric(boolToFloat64(slot.Module.GRC.Receiver.IsTracking), host, slot.Name)
				ch <- c.clkRcvGNSSWarmBoot.mustNewConstMetric(boolToFloat64(slot.Module.GRC.Receiver.IsWarmBooting), host, slot.Name)
				ch <- c.clkRcvGNSSColdBoot.mustNewConstMetric(boolToFloat64(slot.Module.GRC.Receiver.IsColdBooting), host, slot.Name)
			}
		}

		if slot.Module.DCF77 != nil {
			ch <- c.clkRcvDCF77FieldStrength.mustNewConstMetric(slot.Module.DCF77.FieldStrength, host, slot.Name)
			ch <- c.clkRcvDCF77Correlation.mustNewConstMetric(slot.Module.DCF77.Correlation, host, slot.Name)
		}
	}

	c.logger.Debug("Done collecting metrics from Meinberg LTOS device", "target", c.client.Target())
}
