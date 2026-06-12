package entity

type Machine struct {
	CPU  chan CPU  `json:"cpu"`
	RAM  chan RAM  `json:"ram"`
	Disk chan Disk `json:"disk"`
}

func NewMachine() *Machine {
	return &Machine{
		CPU:  make(chan CPU),
		RAM:  make(chan RAM),
		Disk: make(chan Disk),
	}
}

func (m *Machine) Monitoring() {
	
}
