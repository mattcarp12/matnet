package netstack

import (
	"errors"
	"net"
)

type RoutingTable interface {
	Lookup(destination net.IP) route
	AddConnectedRoute(iface NetworkInterface) *route
}

type route struct {
	// Network address of the route
	Network net.IPNet

	// Gateway is the gateway address.
	Gateway net.IP

	// NIC is the network interface.
	Iface NetworkInterface

	// Metric is the metric of the route.
	Metric uint32

	// NextHop is the next hop address.
	NextHop net.IP

	// Connected is true if the route is connected.
	Connected bool
}

type routing_table struct {
	table         []route
	default_route route
}

var ErrNoRouteFound = errors.New("No route found")

func NewRoutingTable() *routing_table {
	return &routing_table{}
}

func (rt *routing_table) add_route(r route) {
	rt.table = append(rt.table, r)
}

func (rt *routing_table) Lookup(destination net.IP) route {
	// Loop through routing table
	for _, r := range rt.table {
		// Check the route to see if it matches the destination
		if r.Network.Contains(destination) {
			if r.Connected {
				// If the route is connected, the next hop is is on
				// the local network, so send the packet directly to it
				r.NextHop = destination
			} else {
				// Otherwise, send the packet to the gateway
				r.NextHop = r.Gateway
			}
			return r
		}
	}

	// Return default route
	return rt.GetDefaultRoute()
}

func (rt *routing_table) GetDefaultRoute() route {
	return rt.default_route
}

func (rt *routing_table) SetDefaultRoute(r route) {
	rt.default_route = r
}

func (rt *routing_table) AddConnectedRoute(iface NetworkInterface) *route {
	// Make route from interface config
	r := route{
		Network: net.IPNet{
			IP:   iface.GetNetworkAddr(),
			Mask: iface.GetNetmask(),
		},
		Gateway:   iface.GetGateway(),
		Iface:     iface,
		Connected: true,
	}
	rt.add_route(r)
	return &r
}
