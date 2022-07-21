package transportlayer

import (
	"encoding/binary"
	"net"

	"github.com/mattcarp12/matnet/netstack"
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
	SrcIP  net.IP
	DstIP  net.IP
	Zero   uint8
	Proto  uint8
	Length uint16
}

func (ph *UdpPsuedoHeader) Marshal() []byte {
	b := []byte{}
	b = append(b, ph.SrcIP...)
	b = append(b, ph.DstIP...)
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
}

func NewUDP() *UdpProtocol {
	udp := &UdpProtocol{
		IProtocol: netstack.NewIProtocol(netstack.ProtocolTypeUDP),
	}
	return udp
}

func (udp *UdpProtocol) HandleRx(skb *netstack.SkBuff) {
	// Create a new UDP header
	h := &UdpHeader{}

	// Unmarshal the UDP header, handle errors
	if err := h.Unmarshal(skb.Data); err != nil {
		skb.Error(err)
		return
	}

	// Set ports on skb
	skb.SetSrcPort(h.SrcPort)
	skb.SetDstPort(h.DstPort)

	// TODO: Handle fragmentation, possibly reassemble

	// Strip the UDP header from the skb
	skb.StripBytes(8)

	// Send to socket layer
	udp.RxUp(skb)
}

func (udp *UdpProtocol) HandleTx(skb *netstack.SkBuff) {
	log.Printf("HandleTx -- UDP packet")

	// setup the UDP header
	h, err := udp.make_udp_header(skb)
	if err != nil {
		skb.Error(err)
		return
	}

	skb.SetSrcPort(h.SrcPort)
	skb.SetDstPort(h.DstPort)
	skb.SetL4Header(h)
	skb.PrependBytes(h.Marshal())

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

	// Set port fields on header
	h.DstPort = skb.GetDstPort()
	h.SrcPort = skb.GetSrcPort()

	// Set length
	h.Length = uint16(len(skb.Data) + 8)

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
	// TODO: Handle IPv6
	p := &UdpPsuedoHeader{
		SrcIP:  skb.GetSrcIP().To4(),
		DstIP:  skb.GetDstIP().To4(),
		Zero:   0,
		Proto:  17,
		Length: h.Length,
	}

	// Create checksum buffer
	b := append(h.Marshal(), skb.Data...)
	b = append(p.Marshal(), b...)

	// Calculate checksum
	h.Checksum = netstack.Checksum(b)

	return nil
}
