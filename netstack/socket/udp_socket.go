package socket

import (
	"net"

	"github.com/mattcarp12/go-net/netstack"
)

/*
	UDP Socket
*/

type udp_socket struct {
	netstack.ISocket
}

func NewUDPSocket(rt netstack.RoutingTable) netstack.Socket {
	s := &udp_socket{
		ISocket: *netstack.NewSocket(),
	}
	s.SetRoutingTable(rt)
	s.SetType(netstack.SocketTypeDatagram)
	s.SetState(netstack.SocketStateClosed)
	return s
}

// Bind...
func (s *udp_socket) Bind(addr netstack.SockAddr) error {
	return nil
}

// Listen...
func (s *udp_socket) Listen(backlog int) error {
	return nil
}

// Accept...
func (s *udp_socket) Accept() (net.Conn, error) {
	return nil, nil
}

// Connect...
func (s *udp_socket) Connect(addr netstack.SockAddr) error {
	return nil
}

// Close...
func (s *udp_socket) Close() error {
	return nil
}

// Read...
func (s *udp_socket) Read(b []byte) (int, error) {
	return 0, nil
}

// Write...
func (s *udp_socket) Write(b []byte) (int, error) {
	return 0, nil
}

// ReadFrom...
func (s *udp_socket) ReadFrom(b []byte, addr *netstack.SockAddr) (int, error) {
	return 0, nil
}

// WriteTo...
func (sock *udp_socket) WriteTo(b []byte, destAddr netstack.SockAddr) (int, error) {
	// Set socket destination address
	sock.SetDestinationAddress(destAddr)

	/// Get route from routing table
	route := sock.GetRoutingTable().Lookup(destAddr.GetIP())

	// Set socket source address
	srcAddr := netstack.SockAddr{}
	srcAddr.IP = route.Iface.GetNetworkAddr()
	sock.SetSourceAddress(srcAddr)

	// Create new skbuff
	skb := netstack.NewSkBuff(b)

	// Set socket on skbuff
	skb.SetSocket(sock)

	// Set skbuff route
	skb.SetRoute(route)

	// Set skbuff type to UDP
	skb.SetType(netstack.ProtocolTypeUDP)

	// Send packet to UDP protocol
	sock.SendSkb(skb)

	// Get response from network stack
	resp := skb.Resp()

	return resp.BytesWritten(), resp.Error()
}
