package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/ltosapi/models"
)

var (
	clkInfo = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"clock_info",
			"Meinberg clock module information as labels (model, serial number, software revision, oscillator type)",
			[]string{"host", "clock_id", "model", "serial_number", "software_revision", "oscillator_type"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkSyncStatus = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"clock_synchronized",
			"Meinberg clock synchronization status (1 = synchronized, 0 = not synchronized)",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkOscillatorWarmedUp = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"clock_oscillator_warmed_up",
			"Meinberg clock oscillator warmed up status (1 = warmed up, 0 = not warmed up)",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkEstTimeQuality = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"clock_estimated_time_quality_seconds",
			"Estimated upper bound in seconds on the time quality of the clock",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
)

func describeClock(ch chan<- *prometheus.Desc) {
	ch <- clkInfo.desc
	ch <- clkSyncStatus.desc
	ch <- clkOscillatorWarmedUp.desc
	ch <- clkEstTimeQuality.desc
}

func (c *Collector) collectClock(ch chan<- prometheus.Metric, host string, slots []models.Slot) {
	for _, slot := range slots {
		if slot.Module == nil || slot.Type != "clk" {
			continue
		}

		oscillatorType := "unknown"
		if slot.Module.SyncStatus != nil {
			oscillatorType = slot.Module.SyncStatus.OscillatorType
			clkSynced := slot.Module.SyncStatus.ClockStatus.Clock == "synchronized"
			oscWarmedUp := slot.Module.SyncStatus.ClockStatus.Oscillator == "warmed-up"

			ch <- clkSyncStatus.mustNewConstMetric(boolToFloat64(clkSynced), host, slot.Name)
			ch <- clkOscillatorWarmedUp.mustNewConstMetric(boolToFloat64(oscWarmedUp), host, slot.Name)
			ch <- clkEstTimeQuality.mustNewConstMetric(slot.Module.SyncStatus.TimeQuality.Seconds(), host, slot.Name)
		}
		ch <- clkInfo.mustNewConstMetric(1.0, host, slot.Name, slot.Module.Info.Model, slot.Module.Info.SerialNumber.String(), slot.Module.Info.SoftwareRevision, oscillatorType)
	}
}
