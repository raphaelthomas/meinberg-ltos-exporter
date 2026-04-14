package collector

import (
	"testing"

	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/ltosapi/models"
)

func TestForEachSlotWithModule(t *testing.T) {
	slots := []models.Slot{
		{Type: models.SlotTypeCPU, Name: "cpu0", Module: &models.SlotModule{}},
		{Type: models.SlotTypeClock, Name: "clk0", Module: &models.SlotModule{}},
		{Type: models.SlotTypeCPU, Name: "cpu1", Module: &models.SlotModule{}},
		{Type: models.SlotTypeClock, Name: "clk1", Module: &models.SlotModule{}},
		{Type: models.SlotTypeCPU, Name: "cpu2", Module: nil},
		{Type: models.SlotTypeClock, Name: "clk2", Module: nil},
	}

	tests := []struct {
		name     string
		slots    []models.Slot
		fn       func([]models.Slot, func(models.Slot))
		expected []string
	}{
		{"multiple cpu slots", slots, forEachCPUSlot, []string{"cpu0", "cpu1"}},
		{"multiple clock slots", slots, forEachClockSlot, []string{"clk0", "clk1"}},
		{"empty input", []models.Slot{}, forEachCPUSlot, nil},
		{"all nil modules", []models.Slot{
			{Type: models.SlotTypeClock, Name: "clk0", Module: nil},
			{Type: models.SlotTypeClock, Name: "clk1", Module: nil},
		}, forEachClockSlot, nil},
		{"no matching type", []models.Slot{
			{Type: models.SlotTypeCPU, Name: "cpu0", Module: &models.SlotModule{}},
		}, forEachClockSlot, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			tt.fn(tt.slots, func(s models.Slot) {
				got = append(got, s.Name)
			})
			if len(got) != len(tt.expected) {
				t.Fatalf("got %v, want %v", got, tt.expected)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("index %d: got %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestBoolToFloat64(t *testing.T) {
	if got := boolToFloat64(true); got != 1.0 {
		t.Errorf("boolToFloat64(true) = %v, want 1.0", got)
	}
	if got := boolToFloat64(false); got != 0.0 {
		t.Errorf("boolToFloat64(false) = %v, want 0.0", got)
	}
}
