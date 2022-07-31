package socket

import (
	"net"

	"github.com/mattcarp12/matnet/netstack"
)

type UDPSocket struct {
	SocketMeta
}

func NewUDPSocket() *UDPSocket {
	s := &UDPSocket{
		SocketMeta: *NewSocketMeta(),
	}
	s.Type = SocketTypeDatagram

	return s
}

// Bind...
func (s *UDPSocket) Bind(addr SockAddr) error {
	return nil
}

// Listen...
func (s *UDPSocket) Listen() error {
	return nil
}

// Accept...
func (s *UDPSocket) Accept() (net.Conn, error) {
	return nil, nil
}

// Connect...
func (s *UDPSocket) Connect(addr SockAddr) error {
	return nil
}

// Close...
func (s *UDPSocket) Close() error {
	// Tell UDP protocol to close socket
	// and delete from socket_manager
	return nil
}

// Read...
func (s *UDPSocket) Read() ([]byte, error) {
	sockLog.Printf("UDP Read()")

	skb := <-s.RxChan

	sockLog.Printf("Read: %v\n", skb)

	return skb.Data, nil
}

// Write...
func (s *UDPSocket) Write(b []byte) (int, error) {
	return 0, nil
}

// ReadFrom...
func (s *UDPSocket) ReadFrom(b []byte, addr *SockAddr) (int, error) {
	return 0, nil
}

// WriteTo...
// At this point the socket should have an iterface and source address set
func (s *UDPSocket) WriteTo(b []byte, destAddr SockAddr) (int, error) {
	// Set socket destination address
	s.DestAddr = destAddr

	// Create new skbuff
	skb := netstack.NewSkBuff(b)

	// Set the skbuff interface
	skb.SetTxIface(s.SocketMeta.Route.Iface)

	// Set the skbuff source and destination addresses
	skb.SetDstAddr(s.DestAddr)
	skb.SetSrcAddr(s.SrcAddr)

	// Set skbuff type to UDP
	skb.SetType(netstack.ProtocolTypeUDP)

	// Send packet to UDP protocol
	s.SocketMeta.Protocol.TxChan() <- skb

	// Wait for response from network stack
	resp := skb.GetResp()

	return resp.BytesWritten, resp.Error
}
