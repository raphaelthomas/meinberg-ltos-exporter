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

type StatusData map[string]any
