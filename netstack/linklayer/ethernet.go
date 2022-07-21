package linklayer

import (
	"encoding/binary"
	"errors"
	logging "log"
	"net"
	"os"

	"github.com/mattcarp12/matnet/netstack"
)

var ethLog = logging.New(os.Stdout, "[Ethernet] ", logging.Ldate|logging.Lmicroseconds|logging.Lshortfile)

// =============================================================================
// EthernetHeader
// =============================================================================

type EthernetAddressPair struct {
	DstAddr net.HardwareAddr
	SrcAddr net.HardwareAddr
}

type EthernetHeader struct {
	addr      EthernetAddressPair
	EtherType uint16
}

// Ethertype values
const (
	EthernetTypeIPv4 = 0x0800
	EthernetTypeARP  = 0x0806
	EthernetTypeIPv6 = 0x86DD
)

const EthernetHeaderSize = 14

var ErrInvalidEthernetHeader = errors.New("invalid ethernet header")

func (eh *EthernetHeader) Unmarshal(b []byte) error {
	if len(b) < EthernetHeaderSize {
		return ErrInvalidEthernetHeader
	}

	// Set HW address in header
	eh.addr.DstAddr = b[0:6]
	eh.addr.SrcAddr = b[6:12]
	eh.EtherType = uint16(b[12])<<8 | uint16(b[13]) // Is in big endian

	return nil
}

func (eh *EthernetHeader) Marshal() []byte {
	d := make([]byte, 14)
	copy(d[0:6], eh.addr.DstAddr)
	copy(d[6:12], eh.addr.SrcAddr)
	binary.BigEndian.PutUint16(d[12:14], eh.EtherType)
	return d
}

func (eh *EthernetHeader) GetType() netstack.ProtocolType {
	return netstack.ProtocolTypeEthernet
}

func (eh *EthernetHeader) GetDstMAC() net.HardwareAddr {
	return eh.addr.DstAddr
}

func (eh *EthernetHeader) GetSrcMAC() net.HardwareAddr {
	return eh.addr.SrcAddr
}

/*
	Helper functions to convert between EtherType and netstack.ProtocolType
*/

func (eh *EthernetHeader) GetL3Type() netstack.ProtocolType {
	switch eh.EtherType {
	case EthernetTypeIPv4:
		return netstack.ProtocolTypeIPv4
	case EthernetTypeARP:
		return netstack.ProtocolTypeARP
	case EthernetTypeIPv6:
		return netstack.ProtocolTypeIPv6
	default:
		return netstack.ProtocolTypeUnknown
	}
}

func GetEtherTypeFromProtocolType(pt netstack.ProtocolType) (uint16, error) {
	switch pt {
	case netstack.ProtocolTypeIPv4:
		return EthernetTypeIPv4, nil
	case netstack.ProtocolTypeARP:
		return EthernetTypeARP, nil
	case netstack.ProtocolTypeIPv6:
		return EthernetTypeIPv6, nil
	default:
		return 0, netstack.ErrProtocolNotFound
	}
}

func IsUnicast(addr net.HardwareAddr) bool {
	return addr[0]&0x01 == 0
}

// =============================================================================
// Ethernet Protocol
// =============================================================================

type EthernetProtocol struct {
	netstack.IProtocol

	// arp represents the neighbor subsystem for either ipv4 (arp) or ipv6 (ndp)
	arp NeighborProtocol
}

func NewEthernet() *EthernetProtocol {
	eth := &EthernetProtocol{
		IProtocol: netstack.NewIProtocol(netstack.ProtocolTypeEthernet),
	}

	return eth
}

func (eth *EthernetProtocol) SetNeighborProtocol(arp NeighborProtocol) {
	eth.arp = arp
}

// HandleRx is called when a skbuff from the network interface is ready
// to be processed by the network stack
func (eth *EthernetProtocol) HandleRx(skb *netstack.SkBuff) {
	// Create empty ethernet header
	ethHdr := EthernetHeader{}

	// Parse the ethernet header
	if err := ethHdr.Unmarshal(skb.Data); err != nil {
		ethLog.Printf("Error parsing ethernet header: %v", err)
		return
	}

	// Get the RxIface
	iface, err := skb.GetRxIface()
	if err != nil {
		ethLog.Printf("Error getting rx iface: %v", err)
		return
	}

	// Check if the packet is destined for this interface
	if IsUnicast(ethHdr.addr.DstAddr) {
		mac := iface.GetHWAddr().String()
		if ethHdr.addr.DstAddr.String() != mac {
			ethLog.Printf("Packet not for this interface (dst: %s, src: %s)", ethHdr.addr.DstAddr.String(), mac)
			return
		}
	} // If multicast or broadcast, continue processing

	// Set L2 fields in the skb
	skb.SetL2Header(&ethHdr)

	// Set skb type to the next layer type (ipv4, arp, etc)
	// by inspecting EtherType field
	skb.SetType(ethHdr.GetL3Type())

	// Strip ethernet header from skb data buffer
	skb.StripBytes(EthernetHeaderSize)

	// If ARP packet, pass to ARP subsystem
	if ethHdr.EtherType == EthernetTypeARP {
		eth.arp.HandleRx(skb)
		return
	}

	// Pass to network layer
	eth.RxUp(skb)
}

// HandleTx is called when a skbuff from the network layer is ready to be sent
func (eth *EthernetProtocol) HandleTx(skb *netstack.SkBuff) {
	// Get the TxIface
	iface, err := skb.GetTxIface()
	if err != nil {
		ethLog.Printf("Error getting tx iface: %s", err.Error())
		return
	}

	/*
		If link layer header is not set, we need to create an ethernet header
		then append it to the skbuff. Then pass the skbuff to network interface.

		In order to fill out the ethernet header, we need to get the
		destination hardware address. We can get this from the arp cache.

		If the arp cache does not have the destination hardware address,
		we need to send an arp request to get it.
	*/

	// Get the destination hardware address from the arp cache
	destHWAddr, err := eth.arp.Resolve(skb.GetDstIP())
	if err != nil {
		ethLog.Printf("Error resolving destination hardware address: %v", err)

		// If we can't get the hardware address, send an arp request
		go eth.arp.SendRequest(skb.GetSrcIP(), skb.GetDstIP(), iface)

		// Make sure to send response for the dropped skb
		// TODO: Make request cache to handle requests once the arp response is received
		skb.Error(err)

		return
	}

	l3header, err := skb.GetL3Header()
	if err != nil {
		ethLog.Printf("Error getting l3 header: %s", err.Error())
		return
	}

	// Create ethernet header
	eth_header := EthernetHeader{}
	eth_header.addr.DstAddr = destHWAddr
	eth_header.addr.SrcAddr = iface.GetHWAddr()
	eth_header.EtherType, err = GetEtherTypeFromProtocolType(l3header.GetType())
	if err != nil {
		ethLog.Printf("Error getting EtherType: %v", err)
		return
	}

	// Prepend ethernet header to skbuff
	skb.PrependBytes(eth_header.Marshal())

	// Pass to network interface
	iface.TxChan() <- skb
}
