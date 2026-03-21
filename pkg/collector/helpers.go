package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/ltosapi/models"
)

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

func forEachSlot(slots []models.Slot, slotType string, fn func(models.Slot)) {
	for _, slot := range slots {
		if slot.Type != slotType || slot.Module == nil {
			continue
		}
		fn(slot)
	}
}

func forEachCPUSlot(slots []models.Slot, fn func(models.Slot)) {
	forEachSlot(slots, "cpu", fn)
}
