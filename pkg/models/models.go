// Package models models the Meinberg LTOS API response data structures.
package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type StatusResponse struct {
	SystemInformation SystemInformation `json:"system-information"`
	Data              StatusData        `json:"data"`
}

type SystemInformation struct {
	Version      string `json:"version"`
	SerialNumber string `json:"serial-number"`
	Hostname     string `json:"hostname"`
	Model        string `json:"model"`
}

type StatusData struct {
	RestAPI      RestAPI          `json:"rest-api"`
	System       System           `json:"system"`
	Notification Notification     `json:"notification"`
	Network      Network          `json:"network"`
	Chassis0     Chassis          `json:"chassis0"`
	NTP          []NTPAssociation `json:"ntp"`
}

type RestAPI struct {
	Version string `json:"api-version"`
}

type System struct {
	UptimeSeconds float64         `json:"uptime"`
	CPULoad       CPULoad         `json:"cpuload"`
	Memory        string          `json:"memory"`
	Storage       []StorageDevice `json:"storage"`
}

type CPULoad struct {
	Load1  float64
	Load5  float64
	Load15 float64
}

// UnmarshalJSON CPULoad of raw form "0.48 0.66 0.57 2/99 25157"
func (c *CPULoad) UnmarshalJSON(data []byte) error {
	var rawCPULoadStr string
	if err := json.Unmarshal(data, &rawCPULoadStr); err != nil {
		return fmt.Errorf("failed to unmarshal CPU load string: %v", err)
	}

	parts := strings.Fields(rawCPULoadStr)
	if len(parts) < 3 {
		return fmt.Errorf("failed to parse CPU load string, expected at least 3 fields: %s", rawCPULoadStr)
	}

	var err error
	c.Load1, err = strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return fmt.Errorf("failed to parse 1-minute CPU load: %v", err)
	}
	c.Load5, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return fmt.Errorf("failed to parse 5-minute CPU load: %v", err)
	}
	c.Load15, err = strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return fmt.Errorf("failed to parse 15-minute CPU load: %v", err)
	}

	return nil
}

type (
	StorageDevice  map[string]any
	Notification   map[string]any
	Network        map[string]any
	Chassis        map[string]any
	NTPAssociation map[string]any
)
