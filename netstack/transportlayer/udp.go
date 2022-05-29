package transportlayer

import (
	"encoding/binary"
	"net"

	"github.com/mattcarp12/go-net/netstack"
)

/*******************************************************************************
	UDP Header
*******************************************************************************/

type UdpHeader struct {
	SrcPort  uint16
	DstPort  uint16
	Length   uint16
	Checksum uint16
}

// Implement netstack.Header interface
func (h *UdpHeader) Marshal() []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint16(b[0:2], h.SrcPort)
	binary.BigEndian.PutUint16(b[2:4], h.DstPort)
	binary.BigEndian.PutUint16(b[4:6], h.Length)
	binary.BigEndian.PutUint16(b[6:8], h.Checksum)
	return b
}

func (h *UdpHeader) Unmarshal(b []byte) error {
	h.SrcPort = binary.BigEndian.Uint16(b[0:2])
	h.DstPort = binary.BigEndian.Uint16(b[2:4])
	h.Length = binary.BigEndian.Uint16(b[4:6])
	h.Checksum = binary.BigEndian.Uint16(b[6:8])
	return nil
}

func (h *UdpHeader) GetType() netstack.ProtocolType {
	return netstack.ProtocolTypeUDP
}

// Implement netstack.L4Header interface
func (h *UdpHeader) GetSrcPort() uint16 {
	return h.SrcPort
}

func (h *UdpHeader) GetDstPort() uint16 {
	return h.DstPort
}

type UdpPsuedoHeader struct {
	SrcAddr net.IP
	DstAddr net.IP
	Zero    uint8
	Proto   uint8
	Length  uint16
}

func (ph *UdpPsuedoHeader) Marshal() []byte {
	b := []byte{}
	b = append(b, ph.SrcAddr...)
	b = append(b, ph.DstAddr...)
	b = append(b, ph.Zero)
	b = append(b, ph.Proto)
	b = append(b, byte(ph.Length>>8))
	b = append(b, byte(ph.Length))
	return b
}

/*******************************************************************************
	UDP Protocol
*******************************************************************************/

type UdpProtocol struct {
	netstack.IProtocol
	port_manager *netstack.PortManager
}

func NewUDP() *UdpProtocol {
	udp := &UdpProtocol{
		IProtocol:    netstack.NewIProtocol(netstack.ProtocolTypeUDP),
		port_manager: netstack.NewPortManager(),
	}
	return udp
}

func (udp *UdpProtocol) HandleRx(skb *netstack.SkBuff) {
	// Create a new UDP header
	// Unmarshal the UDP header, handle errors
	// Handle fragmentation, possibly reassemble
	// Strip the UDP header from the skb
	// Everything is good, send to user space
}

func (udp *UdpProtocol) HandleTx(skb *netstack.SkBuff) {
	log.Printf("HandleTx -- UDP packet")

	// setup the UDP header
	if h, err := udp.make_udp_header(skb); err != nil {
		skb.Error(err)
		return
	} else {
		skb.SetL4Header(h)
		skb.PrependBytes(h.Marshal())
	}

	// Passing to the network layer, so set the skb type
	// to the type of the destination address (ipv4, ipv6, etc)
	if err := set_skb_type(skb); err != nil {
		skb.Error(err)
		return
	}

	// Send to network layer
	udp.TxDown(skb)
}

func (udp *UdpProtocol) make_udp_header(skb *netstack.SkBuff) (*UdpHeader, error) {
	// Create a new UDP header
	h := &UdpHeader{}

	// Set destination port
	h.DstPort = skb.GetSocket().GetDestinationAddress().GetPort()

	// Set source port
	if srcPort, err := udp.port_manager.GetPort(); err != nil {
		return nil, err
	} else {
		h.SrcPort = srcPort
	}

	// Set length
	h.Length = uint16(len(skb.GetBytes()) + 8)

	// Set checksum
	if err := set_udp_checksum(skb, h); err != nil {
		return nil, err
	}

	return h, nil
}

func set_udp_checksum(skb *netstack.SkBuff, h *UdpHeader) error {
	// Set checksum initially to 0
	h.Checksum = 0

	// Make pseudo header
	p := &UdpPsuedoHeader{
		SrcAddr: skb.GetSocket().GetSourceAddress().GetIP(),
		DstAddr: skb.GetSocket().GetDestinationAddress().GetIP(),
		Zero:    0,
		Proto:   17,
		Length:  h.Length,
	}

	// Create checksum buffer
	b := append(h.Marshal(), skb.GetBytes()...)
	b = append(p.Marshal(), b...)

	// Calculate checksum
	h.Checksum = netstack.Checksum(b)

	return nil
}
