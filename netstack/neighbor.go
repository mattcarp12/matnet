package netstack

import "net"

type NeighborProtocol interface {
	Resolve(ip net.IP) (net.HardwareAddr, error)
	SendRequest(ip net.IP, iface NetworkInterface)
}
