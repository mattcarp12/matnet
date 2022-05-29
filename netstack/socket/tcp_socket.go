package socket

import (
	"net"

	"github.com/mattcarp12/go-net/netstack"
)

type tcp_socket struct {
	netstack.ISocket
}

func NewTCPSocket() netstack.Socket {
	s := &tcp_socket{
		ISocket: *netstack.NewSocket(),
	}
	s.SetType(netstack.SocketTypeStream)
	s.SetState(netstack.SocketStateClosed)
	return s
}

// Bind...
func (s *tcp_socket) Bind(addr netstack.SockAddr) error {
	return nil
}

// Listen...
func (s *tcp_socket) Listen() error {
	return nil
}

// Accept...
func (s *tcp_socket) Accept() (net.Conn, error) {
	return nil, nil
}

// Connect...
func (s *tcp_socket) Connect(addr netstack.SockAddr) error {
	return nil
}

// Close...
func (s *tcp_socket) Close() error {
	return nil
}

// Read...
func (s *tcp_socket) Read(b []byte) (int, error) {
	return 0, nil
}

// Write...
func (s *tcp_socket) Write(b []byte) (int, error) {
	return 0, nil
}

// ReadFrom...
func (s *tcp_socket) ReadFrom(b []byte, addr *netstack.SockAddr) (int, error) {
	return 0, nil
}

// WriteTo...
func (s *tcp_socket) WriteTo(b []byte, addr netstack.SockAddr) (int, error) {
	return 0, nil
}
