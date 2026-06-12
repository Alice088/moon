package entity

import "moon/pkg/types"

type RAM struct {
	Usage types.Percent `json:"usage"`
}

func (r *RAM) Collect(machine *Machine) {

}
