// Package models models the Meinberg LTOS API response data structures.
package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	Chassis0     Chassis          `json:"chassis0"`
	NTP          []NTPAssociation `json:"ntp"`
}

type RestAPI struct {
	Version string `json:"api-version"`
}

type System struct {
	UptimeSeconds float64 `json:"uptime"`
	CPULoad       CPULoad `json:"cpuload"`
	Memory        Memory  `json:"memory"`
	Mounts        []Mount `json:"storage"`
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

type Mount struct {
	Size       float64 `json:"size"`
	Used       float64 `json:"used"`
	Mountpoint string  `json:"mountpoint"`
}

func (m *Mount) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Size       float64 `json:"size"`
		Used       float64 `json:"used"`
		Mountpoint string  `json:"mountpoint"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed to unmarshal mount: %v", err)
	}

	m.Size = aux.Size * 1024
	m.Used = aux.Used * 1024
	m.Mountpoint = aux.Mountpoint

	return nil
}

type Notification struct {
	Events []Event `json:"events"`
}

type Event struct {
	Type              string
	Name              string
	LastTriggeredUnix float64
}

func (e *Event) UnmarshalJSON(data []byte) error {
	aux := struct {
		Type          string `json:"type"`
		Name          string `json:"object-id"`
		LastTriggered string `json:"last-triggered"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed to unmarshal event: %v", err)
	}

	e.Type = aux.Type
	e.Name = aux.Name

	if aux.LastTriggered != "never" {
		if parsedTime, err := time.Parse("2006-01-02T15:04:05", aux.LastTriggered); err == nil {
			e.LastTriggeredUnix = float64(parsedTime.Unix())
		}
	}

	return nil
}

type Chassis map[string]any

type LeapIndicator int

// See RFC 5905
const (
	NoWarning              LeapIndicator = iota // no warning
	LastMinuteHas61Seconds                      // last minute of the day has 61 seconds
	LastMinuteHas59Seconds                      // last minute of the day has 59 seconds
	Unknown                                     // unknown (clock unsynchronized)
)

func (l *LeapIndicator) setFromInt(n int) error {
	if n < int(NoWarning) || n > int(Unknown) {
		return fmt.Errorf("invalid leap indicator value: %d", n)
	}
	*l = LeapIndicator(n)
	return nil
}

func (l *LeapIndicator) UnmarshalJSON(data []byte) error {
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		return l.setFromInt(n)
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("failed to unmarshal leap indicator as string: %v", err)
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid leap indicator string value: %q", s)
	}

	return l.setFromInt(n)
}

type UnixFromYYYYMMDDhhmm float64

func (u *UnixFromYYYYMMDDhhmm) UnmarshalJSON(data []byte) error {
	var leapSecRaw string
	if err := json.Unmarshal(data, &leapSecRaw); err != nil {
		return fmt.Errorf("failed to unmarshal leap second string: %v", err)
	}
	leapSecTime, err := time.Parse("200601021504", leapSecRaw)
	if err != nil {
		return fmt.Errorf("failed to parse leap second time: %v", err)
	}

	*u = UnixFromYYYYMMDDhhmm(leapSecTime.Unix())
	return nil
}

type NTPAssociation struct {
	AssociationID  int                  `json:"association-id"`
	Name           string               `json:"object-id"`
	RefID          string               `json:"refid"`
	Stratum        float64              `json:"stratum"`
	Precision      float64              `json:"precision"`
	RootDelay      float64              `json:"rootdelay"`
	RootDispersion float64              `json:"rootdisp"`
	ClockWander    float64              `json:"clk-wander"`
	ClockJitter    float64              `json:"clk-jitter"`
	LeapIndicator  LeapIndicator        `json:"leap"`
	LeapSecondUnix UnixFromYYYYMMDDhhmm `json:"leapsec"`

	// optional peer-only metrics
	Offset     *float64 `json:"offset,omitempty"`
	Delay      *float64 `json:"delay,omitempty"`
	Dispersion *float64 `json:"dispersion,omitempty"`
	Reach      *int     `json:"reach,omitempty"`
}

func (a NTPAssociation) IsSys() bool {
	return a.AssociationID == 0
}
