package socket

import (
	"errors"
	"net"

	"github.com/mattcarp12/matnet/netstack/transportlayer"
)

type TCPSocket struct {
	SocketMeta
}

func NewTCPSocket() *TCPSocket {
	s := &TCPSocket{
		SocketMeta: *NewSocketMeta(),
	}
	s.Type = SocketTypeStream

	return s
}

// Bind...
func (s *TCPSocket) Bind(addr SockAddr) error {
	return nil
}

// Listen...
func (s *TCPSocket) Listen() error {
	return nil
}

// Accept...
func (s *TCPSocket) Accept() (net.Conn, error) {
	return nil, nil
}

// Connect calls the OpenConnection function of the TCP protocol
func (s *TCPSocket) Connect(_ SockAddr) error {
	return s.Protocol.(*transportlayer.TCPProtocol).OpenConnection(s.SocketMeta.SrcAddr, s.SocketMeta.DestAddr, s.SocketMeta.GetNetworkInterface())
}

// Close...
func (s *TCPSocket) Close() error {
	return nil
}

// Read...
func (s *TCPSocket) Read() ([]byte, error) {
	return []byte{}, nil
}

// Write...
func (s *TCPSocket) Write(b []byte) (int, error) {
	return 0, nil
}

// ReadFrom...
func (s *TCPSocket) ReadFrom(b []byte, addr *SockAddr) (int, error) {
	return 0, errors.New("not implemented")
}

// WriteTo...
func (s *TCPSocket) WriteTo(b []byte, addr SockAddr) (int, error) {
	return 0, errors.New("not implemented")
}
