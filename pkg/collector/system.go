package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/ltosapi/models"
)

var (
	buildInfo = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"build_info",
			"Meinberg device build information as labels (e.g., API version, firmware version, host)",
			[]string{"host", "api_version", "firmware_version"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	systemInfo = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"system_info",
			"Meinberg system information as labels (e.g., model, serial number, host)",
			[]string{"host", "model", "serial_number"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	systemCPUInfo = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"system_cpu_info",
			"CPU information as labels (model, serial, etc.)",
			[]string{"host", "model", "serial_number"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	systemUptimeSeconds = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"system_uptime_seconds",
			"System uptime in seconds",
			[]string{"host"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	systemCPULoadAvg = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"system_cpu_load_avg",
			"CPU load averaged over 1, 5, and 15 minutes",
			[]string{"host", "period"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	systemMemoryBytes = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"system_memory_bytes",
			"Total memory in bytes",
			[]string{"host"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	systemMemoryFreeBytes = typedDesc{
		desc: prometheus.NewDesc(
			MetricPrefix+"system_memory_free_bytes",
			"Free memory in bytes",
			[]string{"host"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
)

func describeSystem(ch chan<- *prometheus.Desc) {
	ch <- buildInfo.desc
	ch <- systemInfo.desc
	ch <- systemCPUInfo.desc
	ch <- systemUptimeSeconds.desc
	ch <- systemCPULoadAvg.desc
	ch <- systemMemoryBytes.desc
	ch <- systemMemoryFreeBytes.desc
}

func (c *Collector) collectSystem(ch chan<- prometheus.Metric, host string, systemInformation models.SystemInformation, system models.System, restAPI models.RestAPI, slots []models.Slot) {
	ch <- buildInfo.mustNewConstMetric(1.0, host, restAPI.Version, systemInformation.Version)
	ch <- systemInfo.mustNewConstMetric(1.0, host, systemInformation.Model, systemInformation.SerialNumber.String())
	ch <- systemUptimeSeconds.mustNewConstMetric(system.UptimeSeconds, host)
	ch <- systemCPULoadAvg.mustNewConstMetric(system.CPULoad.Load1, host, "1")
	ch <- systemCPULoadAvg.mustNewConstMetric(system.CPULoad.Load5, host, "5")
	ch <- systemCPULoadAvg.mustNewConstMetric(system.CPULoad.Load15, host, "15")
	ch <- systemMemoryBytes.mustNewConstMetric(system.Memory.Total, host)
	ch <- systemMemoryFreeBytes.mustNewConstMetric(system.Memory.Free, host)

	for _, slot := range slots {
		if slot.Module == nil || slot.Type != "cpu" {
			continue
		}
		ch <- systemCPUInfo.mustNewConstMetric(1.0, host, slot.Module.Info.Model, slot.Module.Info.SerialNumber.String())
	}
}
