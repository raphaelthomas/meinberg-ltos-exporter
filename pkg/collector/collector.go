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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/ltosapi"
)

const (
	MetricNamespace = "meinberg_ltos"
	rootSubsystem   = ""
)

type Config struct {
	System       bool
	Notification bool
	Storage      bool
	Clock        bool
	Receiver     bool
	NTP          bool
}

type Collector struct {
	config Config
	client *ltosapi.Client
	logger *slog.Logger

	up             typedDesc
	scrapeDuration typedDesc
	buildInfo      typedDesc
}

func NewCollector(config Config, client *ltosapi.Client, logger *slog.Logger) *Collector {
	return &Collector{
		config: config,
		client: client,
		logger: logger,
		up: typedDesc{
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(MetricNamespace, rootSubsystem, "up"),
				"Indicates if the Meinberg LTOS device is reachable (1 = up, 0 = down)",
				[]string{"target"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		scrapeDuration: typedDesc{
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(MetricNamespace, rootSubsystem, "scrape_duration_seconds"),
				"Duration of the scrape of the Meinberg LTOS device in seconds",
				[]string{"target"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		buildInfo: typedDesc{
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(MetricNamespace, rootSubsystem, "build_info"),
				"Meinberg device build information as labels (e.g., API version, firmware version, host)",
				[]string{"target", "host", "api_version", "firmware_version"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up.desc
	ch <- c.scrapeDuration.desc
	ch <- c.buildInfo.desc

	if c.config.System {
		describeSystem(ch)
	}
	if c.config.Notification {
		describeNotification(ch)
	}
	if c.config.Storage {
		describeStorage(ch)
	}
	if c.config.Clock {
		describeClock(ch)
	}
	if c.config.Receiver {
		describeReceiverGNSS(ch)
		describeReceiverDCF77(ch)
	}
	if c.config.NTP {
		describeNTP(ch)
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	up := 0.0

	defer func() {
		seconds := time.Since(start).Seconds()
		ch <- c.scrapeDuration.mustNewConstMetric(seconds, c.client.Target())
		ch <- c.up.mustNewConstMetric(up, c.client.Target())
	}()

	c.logger.Debug("Collecting metrics from Meinberg LTOS device", "target", c.client.Target())

	status, err := c.client.FetchStatus()
	if err != nil {
		c.logger.Warn("Failed to fetch Meinberg LTOS device status", "error", err.Error())
		return
	}

	up = 1.0
	host := status.SystemInformation.Hostname
	ch <- c.buildInfo.mustNewConstMetric(1.0, c.client.Target(), host, status.Data.RestAPI.Version, status.SystemInformation.Version)

	if c.config.System {
		c.collectSystem(ch, host, status.SystemInformation, status.Data.System, status.Data.Chassis.Slots)
	}
	if c.config.Notification {
		c.collectNotification(ch, host, status.Data.Notification.Events)
	}
	if c.config.Storage {
		c.collectStorage(ch, host, status.Data.System.Mounts)
	}
	if c.config.NTP {
		c.collectNTP(ch, host, status.Data.NTP)
	}
	if c.config.Clock {
		c.collectClock(ch, host, status.Data.Chassis.Slots)
	}
	if c.config.Receiver {
		c.collectReceiverGNSS(ch, host, status.Data.Chassis.Slots)
		c.collectReceiverDCF77(ch, host, status.Data.Chassis.Slots)
	}

	c.logger.Debug("Done collecting metrics from Meinberg LTOS device", "target", c.client.Target())
}
