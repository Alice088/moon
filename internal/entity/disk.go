package entity

import (
	"moon/pkg/types"
)

type Disk struct {
	Space types.GiB     `json:"space"`
	Usage types.Percent `json:"usage"`
}

func (d *Disk) Collect(machine *Machine) {

}
