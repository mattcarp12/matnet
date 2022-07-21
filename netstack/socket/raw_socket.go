package socket

import (
	"net"
)

type raw_socket struct {
	SocketMeta
}

func NewRawSocket() Socket {
	s := &raw_socket{
		SocketMeta: *NewSocketMeta(),
	}
	s.Type = SocketTypeRaw
	return s
}

// Bind...
func (s *raw_socket) Bind(addr SockAddr) error {
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
func (s *raw_socket) Connect(addr SockAddr) error {
	return nil
}

// Close...
func (s *raw_socket) Close() error {
	return nil
}

// Read...
func (s *raw_socket) Read() ([]byte, error) {
	return []byte{}, nil
}

// Write...
func (s *raw_socket) Write(b []byte) (int, error) {
	return 0, nil
}

// ReadFrom...
func (s *raw_socket) ReadFrom(b []byte, addr *SockAddr) (int, error) {
	return 0, nil
}

// WriteTo...
func (s *raw_socket) WriteTo(b []byte, addr SockAddr) (int, error) {
	return 0, nil
}
