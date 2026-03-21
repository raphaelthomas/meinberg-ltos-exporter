package collector

import (
	"math"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/ltosapi/models"
)

var variableLabelsNTP = []string{"host", "refid", "assoc"}

var (
	ntpStratum = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"ntp_stratum",
			"Meinberg NTP stratum level",
			variableLabelsNTP,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpPrecision = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"ntp_precision_seconds",
			"Meinberg NTP precision in seconds",
			variableLabelsNTP,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpRootDelay = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"ntp_root_delay_seconds",
			"Meinberg NTP root delay in seconds",
			variableLabelsNTP,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpRootDispersion = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"ntp_root_dispersion_seconds",
			"Meinberg NTP root dispersion in seconds",
			variableLabelsNTP,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpClockJitter = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"ntp_clock_jitter_seconds",
			"Meinberg NTP clock jitter in seconds",
			variableLabelsNTP,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpClockWander = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"ntp_clock_wander_seconds_per_second",
			"Meinberg NTP clock wander in seconds per second",
			variableLabelsNTP,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpLeapIndicator = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"ntp_leap_indicator",
			"Meinberg NTP leap indicator (0 = no warning, 1 = last minute has 61 seconds, 2 = last minute has 59 seconds, 3 = unknown)",
			variableLabelsNTP,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpLeapSecond = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"ntp_leap_second_timestamp_seconds",
			"Meinberg NTP leap second (last or next) in seconds since UNIX epoch",
			variableLabelsNTP,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
)

func describeNTP(ch chan<- *prometheus.Desc) {
	ch <- ntpStratum.desc
	ch <- ntpPrecision.desc
	ch <- ntpRootDelay.desc
	ch <- ntpRootDispersion.desc
	ch <- ntpClockJitter.desc
	ch <- ntpClockWander.desc
	ch <- ntpLeapIndicator.desc
	ch <- ntpLeapSecond.desc
}

func (c *Collector) collectNTP(ch chan<- prometheus.Metric, host string, NTPAssocs []models.NTPAssociation) {
	for _, assoc := range NTPAssocs {
		if !assoc.IsSys() {
			continue
		}
		labels := []string{host, assoc.RefID, assoc.Name}

		ch <- ntpStratum.mustNewConstMetric(assoc.Stratum, labels...)
		precisionSeconds := math.Pow(2, assoc.Precision)
		ch <- ntpPrecision.mustNewConstMetric(precisionSeconds, labels...)
		ch <- ntpRootDelay.mustNewConstMetric(assoc.RootDelay, labels...)
		ch <- ntpRootDispersion.mustNewConstMetric(assoc.RootDispersion, labels...)
		ch <- ntpClockJitter.mustNewConstMetric(assoc.ClockJitter, labels...)
		ch <- ntpClockWander.mustNewConstMetric(assoc.ClockWander, labels...)
		ch <- ntpLeapIndicator.mustNewConstMetric(float64(assoc.LeapIndicator), labels...)
		ch <- ntpLeapSecond.mustNewConstMetric(float64(assoc.LeapSecondUnix), labels...)
	}
}
