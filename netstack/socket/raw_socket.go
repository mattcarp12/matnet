package socket

import (
	"net"
)

type RawSocket struct {
	SocketMeta
}

func NewRawSocket() *RawSocket {
	s := &RawSocket{
		SocketMeta: *NewSocketMeta(),
	}
	s.Type = SocketTypeRaw

	return s
}

// Bind...
func (s *RawSocket) Bind(addr SockAddr) error {
	return nil
}

// Listen...
func (s *RawSocket) Listen() error {
	return nil
}

// Accept...
func (s *RawSocket) Accept() (net.Conn, error) {
	return nil, nil
}

// Connect...
func (s *RawSocket) Connect(addr SockAddr) error {
	return nil
}

// Close...
func (s *RawSocket) Close() error {
	return nil
}

// Read...
func (s *RawSocket) Read() ([]byte, error) {
	return []byte{}, nil
}

// Write...
func (s *RawSocket) Write(b []byte) (int, error) {
	return 0, nil
}

// ReadFrom...
func (s *RawSocket) ReadFrom(b []byte, addr *SockAddr) (int, error) {
	return 0, nil
}

// WriteTo...
func (s *RawSocket) WriteTo(b []byte, addr SockAddr) (int, error) {
	return 0, nil
}
