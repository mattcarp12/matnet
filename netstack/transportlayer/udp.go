package transportlayer

import (
	"encoding/binary"
	"net"

	"github.com/mattcarp12/matnet/netstack"
)

// =============================================================================
// UDP Header
// =============================================================================

type UDPHeader struct {
	SrcPort  uint16
	DstPort  uint16
	Length   uint16
	Checksum uint16
}

// Implement netstack.Header interface
func (h *UDPHeader) Marshal() []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint16(b[0:2], h.SrcPort)
	binary.BigEndian.PutUint16(b[2:4], h.DstPort)
	binary.BigEndian.PutUint16(b[4:6], h.Length)
	binary.BigEndian.PutUint16(b[6:8], h.Checksum)
	return b
}

func (h *UDPHeader) Unmarshal(b []byte) error {
	h.SrcPort = binary.BigEndian.Uint16(b[0:2])
	h.DstPort = binary.BigEndian.Uint16(b[2:4])
	h.Length = binary.BigEndian.Uint16(b[4:6])
	h.Checksum = binary.BigEndian.Uint16(b[6:8])
	return nil
}

func (h *UDPHeader) GetType() netstack.ProtocolType {
	return netstack.ProtocolTypeUDP
}

// Implement netstack.L4Header interface
func (h *UDPHeader) GetSrcPort() uint16 {
	return h.SrcPort
}

func (h *UDPHeader) GetDstPort() uint16 {
	return h.DstPort
}

type UDPPsuedoHeader struct {
	SrcIP  net.IP
	DstIP  net.IP
	Zero   uint8
	Proto  uint8
	Length uint16
}

func (ph *UDPPsuedoHeader) Marshal() []byte {
	b := []byte{}
	b = append(b, ph.SrcIP...)
	b = append(b, ph.DstIP...)
	b = append(b, ph.Zero)
	b = append(b, ph.Proto)
	b = append(b, byte(ph.Length>>8))
	b = append(b, byte(ph.Length))
	return b
}

// =============================================================================
// UDP Protocol
// =============================================================================

type UDPProtocol struct {
	netstack.IProtocol
}

func NewUDP() *UDPProtocol {
	udp := &UDPProtocol{
		IProtocol: netstack.NewIProtocol(netstack.ProtocolTypeUDP),
	}
	udp.Log = netstack.NewLogger("UDP")
	return udp
}

func (udp *UDPProtocol) HandleRx(skb *netstack.SkBuff) {
	// Create a new UDP header
	h := &UDPHeader{}

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

func (udp *UDPProtocol) HandleTx(skb *netstack.SkBuff) {
	udp.Log.Printf("HandleTx -- UDP packet")

	// setup the UDP header
	h := udp.makeUDPHeader(skb)

	skb.SetSrcPort(h.SrcPort)
	skb.SetDstPort(h.DstPort)
	skb.SetL4Header(h)
	skb.PrependBytes(h.Marshal())

	// Passing to the network layer, so set the skb type
	// to the type of the destination address (ipv4, ipv6, etc)
	if err := setSkbType(skb); err != nil {
		skb.Error(err)
		return
	}

	// Send to network layer
	udp.TxDown(skb)
}

func (udp *UDPProtocol) makeUDPHeader(skb *netstack.SkBuff) *UDPHeader {
	// Create a new UDP header
	h := &UDPHeader{}

	// Set port fields on header
	h.DstPort = skb.GetDstPort()
	h.SrcPort = skb.GetSrcPort()

	// Set length
	h.Length = uint16(len(skb.Data) + 8)

	// Set checksum
	setUDPChecksum(skb, h)

	return h
}

func setUDPChecksum(skb *netstack.SkBuff, h *UDPHeader) {
	// Set checksum initially to 0
	h.Checksum = 0

	// Make pseudo header
	// TODO: Handle IPv6
	p := &UDPPsuedoHeader{
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
}
