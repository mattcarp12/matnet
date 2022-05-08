package socket

import (
	"net"

	"github.com/mattcarp12/go-net/netstack"
)

type raw_socket struct {
	netstack.ISocket
}

func NewRawSocket() netstack.Socket {
	s := &raw_socket{
		ISocket: *netstack.NewSocket(),
	}
	s.SetType(netstack.SocketTypeRaw)
	s.SetState(netstack.SocketStateClosed)
	return s
}

// Bind...
func (s *raw_socket) Bind(addr netstack.SockAddr) error {
	return nil
}

// Listen...
func (s *raw_socket) Listen() error {
	return nil
}

// Accept...
func (s *raw_socket) Accept() (net.Conn, error) {
	return nil, nil
}

// Connect...
func (s *raw_socket) Connect(addr netstack.SockAddr) error {
	return nil
}

// Close...
func (s *raw_socket) Close() error {
	return nil
}

// Read...
func (s *raw_socket) Read(b []byte) (int, error) {
	return 0, nil
}

// Write...
func (s *raw_socket) Write(b []byte) (int, error) {
	return 0, nil
}

// ReadFrom...
func (s *raw_socket) ReadFrom(b []byte, addr *netstack.SockAddr) (int, error) {
	return 0, nil
}

// WriteTo...
func (s *raw_socket) WriteTo(b []byte, addr netstack.SockAddr) (int, error) {
	return 0, nil
}
