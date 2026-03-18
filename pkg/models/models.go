// Package models models the Meinberg LTOS API response data structures.
package models

import (
	"encoding/json"
	"fmt"
	"regexp"
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
	Memory        Memory          `json:"memory"`
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

type Memory struct {
	Total float64
	Free  float64
}

// UnmarshalJSON memory of raw form "228428 kB total memory, 161732 kB free (70 %)"
func (m *Memory) UnmarshalJSON(data []byte) error {
	var rawMemoryStr string
	if err := json.Unmarshal(data, &rawMemoryStr); err != nil {
		return fmt.Errorf("failed to unmarshal memory string: %v", err)
	}

	totalRe := regexp.MustCompile(`(\d+)\s+kB\s+total`)
	totalMatches := totalRe.FindStringSubmatch(rawMemoryStr)
	if len(totalMatches) < 2 {
		return fmt.Errorf("failed to parse total memory: %q", rawMemoryStr)
	}
	totalMemoryKB, err := strconv.ParseFloat(totalMatches[1], 64)
	if err != nil {
		return fmt.Errorf("failed to parse total memory as float: %v", err)
	}

	freeRe := regexp.MustCompile(`(\d+)\s+kB\s+free`)
	freeMatches := freeRe.FindStringSubmatch(rawMemoryStr)
	if len(freeMatches) < 2 {
		return fmt.Errorf("failed to parse free memory: %q", rawMemoryStr)
	}
	freeMemoryKB, err := strconv.ParseFloat(freeMatches[1], 64)
	if err != nil {
		return fmt.Errorf("failed to parse free memory as float: %v", err)
	}

	m.Total = totalMemoryKB * 1024
	m.Free = freeMemoryKB * 1024

	return nil
}

type (
	StorageDevice  map[string]any
	Notification   map[string]any
	Network        map[string]any
	Chassis        map[string]any
	NTPAssociation map[string]any
)
