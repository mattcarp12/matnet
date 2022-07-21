package netstack

import (
	"net"
)

type RoutingTable interface {
	Lookup(destination net.IP) Route
	AddConnectedRoutes(iface NetworkInterface)
}

type Route struct {
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

type routingTable struct {
	table        []Route
	defaultRoute Route
}

func NewRoutingTable() *routingTable {
	return &routingTable{}
}

func (rt *routingTable) addRoute(r Route) {
	rt.table = append(rt.table, r)
}

func (rt *routingTable) Lookup(destination net.IP) Route {
	// Loop through routing table
	for _, route := range rt.table {
		// Check the route to see if it matches the destination
		if route.Network.Contains(destination) {
			if route.Connected {
				// If the route is connected, the next hop is is on
				// the local network, so send the packet directly to it
				route.NextHop = destination
			} else {
				// Otherwise, send the packet to the gateway
				route.NextHop = route.Gateway
			}

			return route
		}
	}

	// Return default route
	return rt.GetDefaultRoute()
}

func (rt *routingTable) GetDefaultRoute() Route {
	return rt.defaultRoute
}

func (rt *routingTable) AddConnectedRoutes(iface NetworkInterface) {
	// Loop over all IfAddrs of the interface, creating a route for each
	for _, addr := range iface.GetIfAddrs() {
		r := Route{
			Network: net.IPNet{
				IP:   addr.IP,
				Mask: addr.Netmask,
			},
			Gateway:   addr.Gateway,
			Iface:     iface,
			Connected: true,
		}
		rt.addRoute(r)
	}
}

func (rt *routingTable) SetDefaultRoute(net net.IPNet, gateway net.IP, iface NetworkInterface) {
	r := Route{
		Network: net,
		Gateway: gateway,
		Iface:   iface,
	}
	rt.defaultRoute = r
}
