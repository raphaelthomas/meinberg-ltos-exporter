package models

type StatusResponse struct {
	SystemInformation SystemInformation `json:"system-information"`
	Data              StatusData        `json:"data"`
}

type SystemInformation struct {
	Version      string       `json:"version"`
	SerialNumber SerialNumber `json:"serial-number"`
	Hostname     string       `json:"hostname"`
	Model        string       `json:"model"`
}

type StatusData struct {
	RestAPI      RestAPI          `json:"rest-api"`
	System       System           `json:"system"`
	Notification Notification     `json:"notification"`
	Chassis      Chassis          `json:"chassis0"`
	NTP          []NTPAssociation `json:"ntp"`
}

type RestAPI struct {
	Version string `json:"api-version"`
}
