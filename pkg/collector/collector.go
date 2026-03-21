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

const MetricNamespace = "meinberg_ltos"

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

type Collector struct {
	client    *ltosapi.Client
	logger    *slog.Logger
	up        typedDesc
	buildInfo typedDesc
}

func NewCollector(client *ltosapi.Client, logger *slog.Logger) *Collector {
	return &Collector{
		client: client,
		logger: logger,
		up: typedDesc{
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(MetricNamespace, "", "up"),
				"Indicates if the Meinberg LTOS device is reachable (1 = up, 0 = down)",
				[]string{"host", "target"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		buildInfo: typedDesc{
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(MetricNamespace, "", "build_info"),
				"Meinberg device build information as labels (e.g., API version, firmware version, host)",
				[]string{"host", "api_version", "firmware_version"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up.desc
	ch <- c.buildInfo.desc

	describeSystem(ch)
	describeEvent(ch)
	describeStorage(ch)
	describeClock(ch)
	describeReceiverGNSS(ch)
	describeReceiverDCF77(ch)
	describeNTP(ch)
}

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
	ch <- c.buildInfo.mustNewConstMetric(1.0, host, status.Data.RestAPI.Version, status.SystemInformation.Version)

	c.collectSystem(ch, host, status.SystemInformation, status.Data.System, status.Data.Chassis.Slots)
	c.collectEvent(ch, host, status.Data.Notification.Events)
	c.collectStorage(ch, host, status.Data.System.Mounts)
	c.collectNTP(ch, host, status.Data.NTP)
	c.collectClock(ch, host, status.Data.Chassis.Slots)
	c.collectReceiverGNSS(ch, host, status.Data.Chassis.Slots)
	c.collectReceiverDCF77(ch, host, status.Data.Chassis.Slots)

	c.logger.Debug("Done collecting metrics from Meinberg LTOS device", "target", c.client.Target())
}
