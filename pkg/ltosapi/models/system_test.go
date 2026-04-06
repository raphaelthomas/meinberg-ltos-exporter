package models

import (
	"encoding/json"
	"testing"
)

func TestCPULoad_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		load1     float64
		load5     float64
		load15    float64
		expectErr bool
	}{
		{"valid", `"0.48 0.66 0.57 2/99 25157"`, 0.48, 0.66, 0.57, false},
		{"zeros", `"0.00 0.00 0.00 1/50 100"`, 0.0, 0.0, 0.0, false},
		{"too few fields", `"0.48 0.66"`, 0, 0, 0, true},
		{"not a string", `123`, 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c CPULoad
			err := json.Unmarshal([]byte(tt.input), &c)
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.Load1 != tt.load1 || c.Load5 != tt.load5 || c.Load15 != tt.load15 {
				t.Errorf("got {%.2f, %.2f, %.2f}, want {%.2f, %.2f, %.2f}", c.Load1, c.Load5, c.Load15, tt.load1, tt.load5, tt.load15)
			}
		})
	}
}

func TestMemory_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		total     float64
		free      float64
		expectErr bool
	}{
		{"valid", `"228428 kB total memory, 161732 kB free (70 %)"`, 228428 * 1024, 161732 * 1024, false},
		{"no match", `"garbage"`, 0, 0, true},
		{"not a string", `123`, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m Memory
			err := json.Unmarshal([]byte(tt.input), &m)
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if m.Total != tt.total || m.Free != tt.free {
				t.Errorf("got {%.0f, %.0f}, want {%.0f, %.0f}", m.Total, m.Free, tt.total, tt.free)
			}
		})
	}
}

func TestMount_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		size       float64
		used       float64
		mountpoint string
		expectErr  bool
	}{
		{"valid", `{"size":1024,"used":512,"mountpoint":"/data"}`, 1024 * 1024, 512 * 1024, "/data", false},
		{"zeros", `{"size":0,"used":0,"mountpoint":"/"}`, 0, 0, "/", false},
		{"invalid json", `{broken}`, 0, 0, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m Mount
			err := json.Unmarshal([]byte(tt.input), &m)
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if m.Size != tt.size || m.Used != tt.used || m.Mountpoint != tt.mountpoint {
				t.Errorf("got {%.0f, %.0f, %s}, want {%.0f, %.0f, %s}", m.Size, m.Used, m.Mountpoint, tt.size, tt.used, tt.mountpoint)
			}
		})
	}
}
