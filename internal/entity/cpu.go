package entity

import (
	"moon/pkg/types"
	"time"
)

type Usage struct {
	User   types.Percent `json:"user"`
	System types.Percent `json:"system"`
	Idle   types.Percent `json:"idle"`
	IOWait types.Percent `json:"iowait"`
	Steal  types.Percent `json:"steal"`
	Nice   types.Percent `json:"nice"`
}

type CPU struct {
	Usage     []Usage   `json:"usage"`
	Average   Usage     `json:"average"`
	Cores     int       `json:"cores"`
	Model     string    `json:"model"`
	Load1     float64   `json:"load1"`
	Load5     float64   `json:"load5"`
	Load15    float64   `json:"load15"`
	Timestamp time.Time `json:"timestamp"`
}

func (c *CPU) Collect(machine *Machine) {

}
