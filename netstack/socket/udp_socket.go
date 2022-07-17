package socket

import (
	"net"

	"github.com/mattcarp12/go-net/netstack"
)

/*
	UDP Socket
*/

type udp_socket struct {
	netstack.SocketMeta
}

func NewUDPSocket() netstack.Socket {
	s := &udp_socket{
		SocketMeta: *netstack.NewSocketMeta(),
	}
	s.Type = netstack.SocketTypeDatagram
	return s
}

// Bind...
func (s *udp_socket) Bind(addr netstack.SockAddr) error {
	return nil
}

// Listen...
func (s *udp_socket) Listen() error {
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
	// Tell UDP protocol to close socket
	// and delete from socket_manager
	return nil
}

// Read...
func (s *udp_socket) Read() ([]byte, error) {
	// Get skb from RxChan
	skb := <-s.RxChan
	return skb.Data, nil
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
	sock.DestAddr = destAddr

	// Create new skbuff
	skb := netstack.NewSkBuff(b)

	// Set the skbuff interface
	skb.NetworkInterface = sock.SocketMeta.Route.Iface

	// Set skbuff type to UDP
	skb.ProtocolType = netstack.ProtocolTypeUDP

	// Send packet to UDP protocol
	sock.SocketMeta.Protocol.TxChan() <- skb

	// Wait for response from network stack
	resp := skb.GetResp()

	return resp.BytesWritten, resp.Error
}
