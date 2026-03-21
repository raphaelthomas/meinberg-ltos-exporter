package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/ltosapi/models"
)

var (
	clkRcvDCF77FieldStrength = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"clock_receiver_dcf77_field_strength",
			"DCF77 receiver field strength",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkRcvDCF77Correlation = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"clock_receiver_dcf77_correlation",
			"DCF77 receiver correlation",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
)

func describeReceiverDCF77(ch chan<- *prometheus.Desc) {
	ch <- clkRcvDCF77FieldStrength.desc
	ch <- clkRcvDCF77Correlation.desc
}

func (c *Collector) collectReceiverDCF77(ch chan<- prometheus.Metric, host string, slots []models.Slot) {
	for _, slot := range slots {
		if slot.Module == nil || slot.Type != "clk" || slot.Module.DCF77 == nil {
			continue
		}

		ch <- clkRcvDCF77FieldStrength.mustNewConstMetric(slot.Module.DCF77.FieldStrength, host, slot.Name)
		ch <- clkRcvDCF77Correlation.mustNewConstMetric(slot.Module.DCF77.Correlation, host, slot.Name)
	}
}
