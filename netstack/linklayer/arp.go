package linklayer

import (
	"encoding/binary"
	"errors"
	"log"
	"net"
	"os"
	"time"

	"github.com/mattcarp12/matnet/netstack"
)

var arpLog = log.New(os.Stdout, "[ARP] ", log.LstdFlags)

// =============================================================================
// ARP Header
// =============================================================================

var ErrInvalidARPHeader = errors.New("invalid arp header")

const (
	ARP_REQUEST = 1
	ARP_REPLY   = 2
)

const (
	ARP_HardwareTypeEthernet = 1
	ARP_ProtocolTypeIPv4     = 0x0800
)

type ARPHeader struct {
	HardwareType uint16
	ProtocolType uint16
	HardwareSize uint8
	ProtocolSize uint8
	OpCode       uint16
	SourceHWAddr net.HardwareAddr
	SourceIPAddr net.IP
	TargetHWAddr net.HardwareAddr
	TargetIPAddr net.IP
}

func (ah *ARPHeader) Unmarshal(b []byte) error {
	if len(b) < 8 {
		return ErrInvalidARPHeader
	}

	ah.HardwareType = binary.BigEndian.Uint16(b[0:2])
	ah.ProtocolType = binary.BigEndian.Uint16(b[2:4])
	ah.HardwareSize = b[4]
	ah.ProtocolSize = b[5]
	ah.OpCode = binary.BigEndian.Uint16(b[6:8])

	// Parse variable length addresses
	minLen := 8 + int(ah.HardwareSize)*2 + int(ah.ProtocolSize)*2
	if len(b) < minLen {
		return ErrInvalidARPHeader
	}

	// Source HW address
	ah.SourceHWAddr = b[8 : 8+int(ah.HardwareSize)]

	// Source IP address
	ah.SourceIPAddr = b[8+int(ah.HardwareSize) : 8+int(ah.HardwareSize)+int(ah.ProtocolSize)]

	// Target HW address
	ah.TargetHWAddr = b[8+int(ah.HardwareSize)+int(ah.ProtocolSize) : 8+int(ah.HardwareSize)*2+int(ah.ProtocolSize)]

	// Target IP address
	ah.TargetIPAddr = b[8+int(ah.HardwareSize)*2+int(ah.ProtocolSize) : 8+int(ah.HardwareSize)*2+int(ah.ProtocolSize)*2]

	return nil
}

func (arpHeader *ARPHeader) Marshal() []byte {
	b := make([]byte, 8+len(arpHeader.SourceHWAddr)+len(arpHeader.SourceIPAddr)+len(arpHeader.TargetHWAddr)+len(arpHeader.TargetIPAddr))

	// Hardware type
	binary.BigEndian.PutUint16(b[0:2], arpHeader.HardwareType)

	// Protocol type
	binary.BigEndian.PutUint16(b[2:4], arpHeader.ProtocolType)

	// Hardware size
	b[4] = arpHeader.HardwareSize

	// Protocol size
	b[5] = arpHeader.ProtocolSize

	// Op code
	binary.BigEndian.PutUint16(b[6:8], uint16(arpHeader.OpCode))

	// Source HW address
	copy(b[8:8+len(arpHeader.SourceHWAddr)], arpHeader.SourceHWAddr)

	// Source IP address
	copy(b[8+len(arpHeader.SourceHWAddr):8+len(arpHeader.SourceHWAddr)+len(arpHeader.SourceIPAddr)], arpHeader.SourceIPAddr)

	// Target HW address
	copy(b[8+len(arpHeader.SourceHWAddr)+len(arpHeader.SourceIPAddr):8+len(arpHeader.SourceHWAddr)+len(arpHeader.SourceIPAddr)+len(arpHeader.TargetHWAddr)], arpHeader.TargetHWAddr)

	// Target IP address
	copy(b[8+len(arpHeader.SourceHWAddr)+len(arpHeader.SourceIPAddr)+len(arpHeader.TargetHWAddr):8+len(arpHeader.SourceHWAddr)+len(arpHeader.SourceIPAddr)+len(arpHeader.TargetHWAddr)+len(arpHeader.TargetIPAddr)], arpHeader.TargetIPAddr)

	return b
}

func (arpHeader *ARPHeader) GetDstIP() net.IP {
	return arpHeader.TargetIPAddr
}

func (arpHeader *ARPHeader) GetSrcIP() net.IP {
	return arpHeader.SourceIPAddr
}

func (arpHeader *ARPHeader) GetType() netstack.ProtocolType {
	return netstack.ProtocolTypeARP
}

func (arpHeader *ARPHeader) GetL4Type() netstack.ProtocolType {
	// This shouldn't be needed...
	return netstack.ProtocolTypeUnknown
}

// =============================================================================
// ARP Protocol
// =============================================================================

type ARPProtocol struct {
	netstack.IProtocol
	cache *ARPCache
}

func NewARP() *ARPProtocol {
	return &ARPProtocol{
		IProtocol: netstack.NewIProtocol(netstack.ProtocolTypeARP),
		cache:     NewARPCache(),
	}
}

func (arp *ARPProtocol) HandleRx(skb *netstack.SkBuff) {
	arpLog.Printf("Received ARP packet")
	// Create empty arp header
	arpHeader := &ARPHeader{}

	// parse the arp header
	if err := arpHeader.Unmarshal(skb.Data); err != nil {
		log.Printf("Error parsing arp header: %v", err)
		return
	}

	// Check if arp Hardware type is Ethernet
	if arpHeader.HardwareType != ARP_HardwareTypeEthernet {
		log.Printf("Unsupported hardware type: %v", arpHeader.HardwareType)
		return
	}

	// Check if arp protocol type is IPv4
	if arpHeader.ProtocolType != ARP_ProtocolTypeIPv4 {
		log.Printf("Unsupported protocol type: %v", arpHeader.ProtocolType)
		return
	}

	// Update the arp cache with the entry for source ip
	arp.cache.Update(arpHeader)

	// Check if this is an arp request
	if arpHeader.OpCode != ARP_REQUEST {
		return
	}

	// Make sure TargetIP equals our IP
	rxIface, err := skb.GetRxIface()
	if err != nil {
		log.Printf("Error getting rx iface: %s", err.Error())
		return
	}

	if !rxIface.HasIPAddr(arpHeader.TargetIPAddr) {
		log.Printf("ARP request not for our IP: %v", arpHeader.TargetIPAddr)
		return
	}

	arp.ARPReply(arpHeader, rxIface)
}

func (arp *ARPProtocol) ARPReply(inArpHeader *ARPHeader, iface netstack.NetworkInterface) {
	// Create a new arp header from the request header
	arpReplyHeader := &ARPHeader{}
	arpReplyHeader.HardwareType = inArpHeader.HardwareType
	arpReplyHeader.ProtocolType = inArpHeader.ProtocolType
	arpReplyHeader.HardwareSize = inArpHeader.HardwareSize
	arpReplyHeader.ProtocolSize = inArpHeader.ProtocolSize
	arpReplyHeader.OpCode = ARP_REPLY
	arpReplyHeader.TargetHWAddr = inArpHeader.SourceHWAddr
	arpReplyHeader.TargetIPAddr = inArpHeader.SourceIPAddr
	arpReplyHeader.SourceHWAddr = iface.GetHWAddr()
	arpReplyHeader.SourceIPAddr = inArpHeader.TargetIPAddr

	// Create a new arp skb
	arpReplySkb := netstack.NewSkBuff(arpReplyHeader.Marshal())

	// Set the network interface to the same interface the request came from
	arpReplySkb.SetTxIface(iface)

	// Set the type of the skb to the link layer type (ethernet, etc),
	// which we get from the network interface
	arpReplySkb.SetType(iface.GetType())

	// Set L3 header in the skb
	arpReplySkb.SetL3Header(arpReplyHeader)

	// Send the arp reply down to link layer
	// arp.TxDown(arpReplySkb)
	// TODO: Find a better way to do this.
	arp.GetLayer().TxChan() <- arpReplySkb

	// Get the skb response
	arpReplySkb.GetResp()
}

func (arp *ARPProtocol) ARPRequest(srcIP, targetIP net.IP, iface netstack.NetworkInterface) {
	// Create a new arp header
	arpRequestHeader := &ARPHeader{}
	arpRequestHeader.HardwareType = ARP_HardwareTypeEthernet
	arpRequestHeader.ProtocolType = ARP_ProtocolTypeIPv4
	arpRequestHeader.HardwareSize = 6
	arpRequestHeader.ProtocolSize = 4
	arpRequestHeader.OpCode = ARP_REQUEST
	arpRequestHeader.SourceHWAddr = iface.GetHWAddr()
	arpRequestHeader.SourceIPAddr = srcIP.To4()
	arpRequestHeader.TargetIPAddr = targetIP.To4()
	arpRequestHeader.TargetHWAddr = net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff} // Broadcast address

	// Since we're making an ARP request for the target IP, we need to
	// set the target hardware address to the broadcast address (ff:ff:ff:ff:ff:ff)
	// in the arp cache.
	arp.cache.Put(targetIP, net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff})

	// Create a new arp skb
	rawArpHeader := arpRequestHeader.Marshal()
	arpSkb := netstack.NewSkBuff(rawArpHeader)

	// Set src and dest addresses in the skb
	arpSkb.SetL3Header(arpRequestHeader)
	arpSkb.SetSrcIP(srcIP)
	arpSkb.SetDstIP(targetIP)

	// Set the network interface
	arpSkb.SetTxIface(iface)

	// Set the type of the skb to the link layer type (ethernet, etc),
	// which we get from the network interface
	arpSkb.SetType(iface.GetType())

	// Send the arp request down to link layer
	// arp.TxDown(arpSkb)
	// TODO: Figure out a better way to do this.
	arp.GetLayer().TxChan() <- arpSkb

	// Get the skb response
	skbResp := arpSkb.GetResp()

	// log the response
	arpLog.Printf("ARP Request SkbResponse is: %+v", skbResp)
}

// This is not used, use ARPRequest instead
func (arp *ARPProtocol) HandleTx(skb *netstack.SkBuff) {}

func (arp *ARPProtocol) Resolve(ip net.IP) (net.HardwareAddr, error) {
	return arp.cache.Lookup(ip)
}

// ==============================================================================
// ARP Cache
// ==============================================================================

// TODO: Start goroutine to clean up cache periodically

type ARPCacheEntry struct {
	MAC       net.HardwareAddr
	timestamp time.Time
}

type ARPCache map[string]ARPCacheEntry

const ARPTimeout = 5

func NewARPCache() *ARPCache {
	return &ARPCache{}
}

func (c *ARPCache) Update(h *ARPHeader) {
	ip := h.SourceIPAddr.String()

	(*c)[ip] = ARPCacheEntry{
		MAC:       h.SourceHWAddr,
		timestamp: time.Now(),
	}
}

func (c *ARPCache) Cleanup() {
	now := time.Now()
	for ip, entry := range *c {
		if now.Sub(entry.timestamp) > ARPTimeout*time.Second {
			delete(*c, ip)
		}
	}
}

var ErrArpCacheMiss error = errors.New("arp cache miss")

func (c *ARPCache) Lookup(ip net.IP) (net.HardwareAddr, error) {
	if entry, ok := (*c)[ip.String()]; ok {
		return entry.MAC, nil
	}

	return nil, ErrArpCacheMiss
}

func (c *ARPCache) Put(ip net.IP, mac net.HardwareAddr) {
	(*c)[ip.String()] = ARPCacheEntry{
		MAC:       mac,
		timestamp: time.Now(),
	}
}
