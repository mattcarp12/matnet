package socket

import (
	"errors"
	"fmt"
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
	tcpProtocol, ok := s.Protocol.(*transportlayer.TCPProtocol)
	if !ok {
		return errors.New("TCP socket does not have a TCP protocol")
	}

	err := tcpProtocol.OpenConnection(
		s.SocketMeta.SrcAddr,
		s.SocketMeta.DestAddr,
		s.SocketMeta.GetNetworkInterface(),
	)
	if err != nil {
		return fmt.Errorf("TCPSocket Connect: error opening connection: %v", err)
	}

	return nil
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
