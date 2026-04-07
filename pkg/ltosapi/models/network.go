package models

type Network struct {
	Ports []Port `json:"ports"`
}

type Port struct {
	Name       string `json:"object-id"`
	Link       bool   `json:"link"`
	Speed      string `json:"speed"`
	Duplex     string `json:"duplex"`
	MACAddress string `json:"mac-address"`
	CardName   string `json:"card-name"`

	Statistics *PortStatistics `json:"statistics"`
}

type PortStatistics struct {
	RxBytes   float64 `json:"rx-bytes"`
	TxBytes   float64 `json:"tx-bytes"`
	RxPackets float64 `json:"rx-packets"`
	TxPackets float64 `json:"tx-packets"`
	RxErrors  float64 `json:"rx-errors"`
	TxErrors  float64 `json:"tx-errors"`
	RxDropped float64 `json:"rx-dropped"`
	TxDropped float64 `json:"tx-dropped"`
}
