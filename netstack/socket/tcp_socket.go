package socket

import (
	"errors"
	"net"

	"github.com/mattcarp12/matnet/netstack/transportlayer"
)

type tcp_socket struct {
	SocketMeta
}

func NewTCPSocket() Socket {
	s := &tcp_socket{
		SocketMeta: *NewSocketMeta(),
	}
	s.Type = SocketTypeStream
	return s
}

// Bind...
func (s *tcp_socket) Bind(addr SockAddr) error {
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

// Connect calls the OpenConnection function of the TCP protocol
func (s *tcp_socket) Connect(_ SockAddr) error {
	return s.Protocol.(*transportlayer.TcpProtocol).OpenConnection(s.SocketMeta.SrcAddr, s.SocketMeta.DestAddr)
}

// Close...
func (s *tcp_socket) Close() error {
	return nil
}

// Read...
func (s *tcp_socket) Read() ([]byte, error) {
	return []byte{}, nil
}

// Write...
func (s *tcp_socket) Write(b []byte) (int, error) {
	return 0, nil
}

// ReadFrom...
func (s *tcp_socket) ReadFrom(b []byte, addr *SockAddr) (int, error) {
	return 0, errors.New("not implemented")
}

// WriteTo...
func (s *tcp_socket) WriteTo(b []byte, addr SockAddr) (int, error) {
	return 0, errors.New("not implemented")
}
