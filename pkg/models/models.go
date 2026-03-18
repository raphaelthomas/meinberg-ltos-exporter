// Package models models the Meinberg LTOS API response data structures.
package models

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

type (
	System         map[string]any
	Notification   map[string]any
	Network        map[string]any
	Chassis        map[string]any
	NTPAssociation map[string]any
)
