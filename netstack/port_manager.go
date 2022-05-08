package netstack

import "errors"

type PortManager struct {
	current_port   uint16
	assigned_ports map[uint16]bool
}

func NewPortManager() *PortManager {
	return &PortManager{current_port: 40000, assigned_ports: make(map[uint16]bool)}
}

func (pm *PortManager) GetPort() (uint16, error) {
	for i := pm.current_port; i < 65535; i++ {
		if !pm.assigned_ports[i] {
			pm.assigned_ports[i] = true
			pm.current_port = i
			return i, nil
		}
	}
	return 0, errors.New("no ports available")
}
